---
title: "Manager"
draft: true
weight: 35
---

controller-runtimeでは、複数のコントローラを管理するための機能としてmanagerを提供しています。
すでにこれまでmanagerのいくつかの機能を紹介してきましたが、他にもたくさんの便利な機能を持っています。


## Leader Election



## Runnable

カスタムコントローラの実装において、Reconcile Loop以外にもgoroutineを立ち上げて定期的に実行したり、
何らかのイベントを待ち受けたりしたい場合があります。
managerではそのような処理を実現するためにRunnableの仕組みを提供しています。

例えばTopoLVMでは、定期的なメトリクスの収集やgRPCサーバの起動用にRunnableを利用しています。

- https://github.com/cybozu-go/topolvm/tree/master/runners

さて、Runnable機能を利用するためには、まず[Runnable](https://pkg.go.dev/sigs.k8s.io/controller-runtime/pkg/manager?tab=doc#Runnable)
インタフェースを実装した以下のようなコードを用意します。

{{% code file="/static/codes/tenant/runners/runner.go" language="go" %}}

StartメソッドはmanagerのStartを呼び出した際に、goroutineとして呼び出されます。
引数のchによりmanagerからの終了通知を受け取ることができます。

```go
err = mgr.Add(&runners.Runner{})
```

また、[LeaderElectionRunnable](https://pkg.go.dev/sigs.k8s.io/controller-runtime/pkg/manager?tab=doc#LeaderElectionRunnable)
インタフェースを実装し、NeedLeaderElectionメソッドがtrueを返すようにすると、
リーダーとして動作しているときににだけStartメソッドが実行されるようになります。

## recorderProvider

カスタムリソースのStatusには、現在の状態だけが保存されています。
これまでどのような処理が実施されてきたのかを確認したい場合

イベントを記録するためのrecorderオブジェクトは以下のように取得することができます。これをReconcilerに渡して利用します。

```go
recorder := mgr.GetEventRecorderFor("tenant-controller")
```



```go
// Reconcileによる更新処理が成功した場合に、Normalタイプのイベントを記録
r.Recorder.Event(&tenant, corev1.EventTypeNormal, "Updated", "the tenant was updated")

// Reconcileによる処理が失敗した場合に、Warningタイプのイベントを記録
r.Recorder.Eventf(&tenant, corev1.EventTypeWarning, "Failed", "failed to reconciled: %s", err.Error())
```

なお、namespace-scopeのカスタムリソースの場合は、カスタムリソースと同じnamespaceにEventリソースが作成されますが、cluster-scopeのカスタムリソースの場合は、default namespaceにEventリソースが作成されます。

そのため、下記のようなRoleとRoleBindingを用意して、コントローラがdefault namespaceにEventリソースを作成できるように設定しておきましょう。

{{% code file="/static/codes/tenant/config/rbac/event_recorder_rbac.yaml" language="yaml" %}}

## metricsListener

```go
	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                 scheme,
		MetricsBindAddress:     metricsAddr,
		Port:                   9443,
		LeaderElection:         enableLeaderElection,
		LeaderElectionID:       "27475f02.example.com",
		HealthProbeBindAddress: probeAddr,
	})
```

これだけで、CPUやメモリの使用量などの基本的なメトリクスと、Reconcileにかかった時間やKubernetesクライアントのレイテンシーなど、controller-runtime関連のメトリクスが収集できるようになります。

さらに追加でコントローラ固有のメトリクスを収集したい場合は、以下のようなコードを記述します。
詳しくは[Prometheusのドキュメント](https://prometheus.io/docs/instrumenting/writing_exporters/)を参照してください。

{{% code file="/static/codes/tenant/controllers/metrics.go" language="go" %}}

Reconcile処理の中で以下の処理を呼び出します。

```go
// namespaceが追加されたときに以下の処理を実行
addedNamespaces.Inc()

// namespaceが削除されたときに以下の処理を実行
removedNamespaces.Inc()
```



## healthProbeListener


## Inject

managerには、ReconcilerやRunnerなどに特定のオブジェクトをインジェクトする機能があります。
InjectClientやInjectLoggerなどのインタフェースを実装すると、managerのStartメソッドを実行したタイミングで、
Kubernetesクライアントやロガーのオブジェクトを受け取ることが可能です。

利用可能なインタフェースについては下記のパッケージを参照してください。
- [inject](https://pkg.go.dev/sigs.k8s.io/controller-runtime/pkg/runtime/inject?tab=doc)

これらのオブジェクトのほとんどは、Getメソッドで取得することができるため、injectの使いみちはそれほど多くありません。
しかし、StopChannelだけはinjectでしか取得することができないため、Reconcilerの中でmanagerからの
終了通知を受け取りたい場合は、以下のようにInjectStopChannelを利用することになります。

{{% code file="/static/codes/tenant/controllers/inject.go" language="go" %}}

このStopChannelから`context.Context`を作成しておけば、Reconcile Loopの中でmanagerからの終了通知を受け取れます。

```go
type TenantReconciler struct {
	stopCh   <-chan struct{}
}

func (r *TenantReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := contextFromStopChannel(r.stopCh)
```

ただし、Reconcileは短時間で終了することが望ましいとされていますので、
長時間ブロックするような処理はなるべく記述しないようにしましょう。
