# Manager

[Manager](https://pkg.go.dev/sigs.k8s.io/controller-runtime/pkg/manager?tab=doc#Manager)は、複数のコントローラを管理し、リーダー選出機能や、メトリクスやヘルスチェックサーバーとしての機能などを提供します。

すでにこれまでManagerのいくつかの機能を紹介してきましたが、他にもたくさんの便利な機能を持ってるのでここで紹介していきます。

## Leader Election

カスタムコントローラの可用性を向上させたい場合、Deploymentの機能を利用してカスタムコントローラのPodを複数個立ち上げます。
しかし、Reconcile処理が同じリソースに対して何らかの処理を実行した場合、競合が発生してしまうかもしれません。

そこで、Managerはリーダー選出機能を提供しています。
これにより複数のプロセスの中から1つだけリーダーを選出し、リーダーに選ばれたプロセスだけがReconcile処理を実行できるようになります。

リーダー選出の利用方法は非常に簡単で、`NewManager`のオプションの`LeaderElection`にtrueを指定し、`LeaderElectionID`にリーダー選出用のIDを指定するだけです。
リーダー選出は、同じ`LeaderElectionID`を指定したプロセスの中から一つだけリーダーを選ぶという挙動になります。

[import:"new-manager",unindent:"true"](../../codes/tenant/main.go)

それでは、[config/manager/manager.yaml](../../codes/tenant/config/manager/manager.yaml)の`replicas`フィールドを2に変更して、テナントコントローラをデプロイしてみましょう。

デプロイされた2つのPodのログを表示させてみると、リーダーに選出された方のPodだけがReconcile処理をおこなっている様子が確認できると思います。

リーダー選出の機能にはConfigMapが利用されています。
下記のようにConfigMapを表示させてみると、`metadata.annotations["control-plane.alpha.kubernetes.io/leader"]`に、現在のリーダーの情報が保存されていることがわかります。

```
$ kubectl get -n tenant-system configmap 27475f02.example.com -o yaml
apiVersion: v1
kind: ConfigMap
metadata:
  annotations:
    control-plane.alpha.kubernetes.io/leader: '{"holderIdentity":"tenant-controller-manager-5d6f8bbd95-h5jpx_85d3882f-1419-42dc-928b-bd7d7dfb8cff","leaseDurationSeconds":15,"acquireTime":"2020-07-25T07:10:29Z","renewTime":"2020-07-25T10:31:41Z","leaderTransitions":10}'
  creationTimestamp: "2020-07-18T09:00:57Z"
  name: 27475f02.example.com
  namespace: tenant-system
  resourceVersion: "1206094"
  selfLink: /api/v1/namespaces/tenant-system/configmaps/27475f02.example.com
  uid: bb91b084-8c8e-4361-9454-071930a1d67c
```

なお、Admission Webhook処理は競合の心配がないため、リーダーではないプロセスの場合でも呼び出されます。

## Runnable

カスタムコントローラの実装において、Reconcile Loop以外にもgoroutineを立ち上げて定期的に実行したり、何らかのイベントを待ち受けたりしたい場合があります。
Managerではそのような処理を実現するための仕組みを提供しています。

例えばTopoLVMでは、定期的なメトリクスの収集やgRPCサーバの起動用にRunnableを利用しています。

- [https://github.com/topolvm/topolvm/tree/master/runners](https://github.com/topolvm/topolvm/tree/master/runners)

Runnable機能を利用するためには、まず[Runnable](https://pkg.go.dev/sigs.k8s.io/controller-runtime/pkg/manager?tab=doc#Runnable)インタフェースを実装した以下のようなコードを用意します。

[import, title="runner.go"](../../codes/tenant/runners/runner.go)

StartメソッドはmanagerのStartを呼び出した際に、goroutineとして呼び出されます。
引数のchによりmanagerからの終了通知を受け取ることができます。

```go
err = mgr.Add(&runners.Runner{})
```

`Runnable` インタフェースを実装しただけだと、リーダーとして動作している Manager でしか動かないようになります。
リーダーでなくても常時動かしたい処理である場合、[LeaderElectionRunnable](https://pkg.go.dev/sigs.k8s.io/controller-runtime/pkg/manager?tab=doc#LeaderElectionRunnable)インタフェースを実装し、NeedLeaderElectionメソッドで `false` を返すようにします。

## recorderProvider

カスタムリソースのStatusには、現在の状態が保存されています。
一方、これまでどのような処理が実施されてきたのかを記録したい場合、Kubernetesが提供する[Event](https://pkg.go.dev/k8s.io/api/core/v1?tab=doc#Event)リソースを利用することができます。

Managerはイベントを記録するための機能を提供しており、以下のように取得することができます。

```go
recorder := mgr.GetEventRecorderFor("tenant-controller")
```

この[EventRecorder](https://pkg.go.dev/k8s.io/client-go/tools/record?tab=doc#EventRecorder)をReconcilerに渡して利用します。

Eventを記録するための関数として、`Event`, `Eventf`, `AnnotatedEventf`などが用意されており、下記のように利用することができます。
なお、イベントタイプには`EventTypeNormal`, `EventTypeWarning`のみ指定することができます。

```go
// Reconcileによる更新処理が成功した場合に、Normalタイプのイベントを記録
r.Recorder.Event(&tenant, corev1.EventTypeNormal, "Updated", "the tenant was updated")

// Reconcileによる処理が失敗した場合に、Warningタイプのイベントを記録
r.Recorder.Eventf(&tenant, corev1.EventTypeWarning, "Failed", "failed to reconciled: %s", err.Error())
```

このEventリソースは第1引数で指定したリソースに結びいており、namespace-scopedリソースの場合はそのリソースと同じnamespaceにEventリソースが作成されます。
一方cluster-scopedリソースの場合は、default namespaceにEventリソースが作成されます。

テナントリソースはcluster-scopedリソースなのでEventはdefault namespaceに作成されます。
そこで下記のようなRoleとRoleBindingを用意して、テナントコントローラがdefault namespaceにEventリソースを作成できるように設定しておきましょう。

[import, title="event_recorder_rbac.yaml"](../../codes/tenant/config/rbac/event_recorder_rbac.yaml)

それでは、作成されたEventリソースを確認してみましょう。なお、Eventリソースはデフォルトで1時間経つと消えてしまいます。

```
$ kubectl get events -n default
LAST SEEN   TYPE     REASON    OBJECT                 MESSAGE
6s          Normal   Updated   tenant/tenant-sample   the tenant was updated
```

## healthProbeListener

Managerには、ヘルスチェック用のAPIのエンドポイントを作成する機能が用意されています。

まずは、Managerの作成時に`HealthProbeBindAddress`でエンドポイントのアドレスを指定します。

[import:"new-manager",unindent:"true"](../../codes/tenant/main.go)

そして、`AddHealthzCheck`と`AddReadyzCheck`で、ハンドラの登録をおこないます。
ここでは`healthz.Ping`という何もしない関数を利用していますが、独自の関数を登録することも可能です。

[import:"health",unindent:"true"](../../codes/tenant/main.go)

これでコントローラにヘルスチェック用のAPIが実装できました。
マニフェストに`livenessProbe`や`readinessProbe`の設定を追加しておきましょう。

[import:"probe"](../../codes/tenant/config/manager/manager.yaml)

## Inject

Managerには、ReconcilerやRunnerなどに特定のオブジェクトをインジェクトする機能があります。
InjectClientやInjectLoggerなどのインタフェースを実装すると、managerのStartメソッドを実行したタイミングで、
Kubernetesクライアントやロガーのオブジェクトを受け取ることが可能です。

利用可能なインタフェースについては下記のパッケージを参照してください。
- [inject](https://pkg.go.dev/sigs.k8s.io/controller-runtime/pkg/runtime/inject?tab=doc)

これらのオブジェクトのほとんどは、Getメソッドで取得することができるため、injectの使いみちはそれほど多くありません。
しかし、StopChannelだけはinjectでしか取得することができないため、Reconcilerの中でmanagerからの
終了通知を受け取りたい場合は、以下のようにInjectStopChannelを利用することになります。

[import, title="inject.go"](../../codes/tenant/controllers/inject.go)

このStopChannelから`context.Context`を作成しておけば、Reconcile Loopの中でmanagerからの終了通知を受け取れます。

```go
type TenantReconciler struct {
	stopCh   <-chan struct{}
}

func (r *TenantReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := contextFromStopChannel(r.stopCh)
```

ただし、Reconcileは短時間で終了することが望ましいとされていますので、長時間ブロックするような処理はなるべく記述しないようにしましょう。
