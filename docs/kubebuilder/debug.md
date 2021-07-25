# 手軽な動作確認

## make runによる実行

コントローラの開発中は何度もプログラムをビルドして実行し直すことになりますが、コンテナイメージのビルドに時間がかかるため非効率になりがちです。
また、コンテナとして動作しているプログラムはデバッグ実行がしにくいという問題もあります。

Kubebuilderによって生成されたMakefileには`make run`というターゲットが用意されており、コントローラをローカルのプロセスとして動作させることが可能です。

ただし、Kubernetesクラスタ上で動作させる場合といくつか違いがあるので、注意して利用してください。

* リーダーエレクション機能を利用する場合は、[Options](https://pkg.go.dev/sigs.k8s.io/controller-runtime/pkg/manager?tab=doc#Options)の`LeaderElectionNamespace`を指定する必要があります。
* 証明書やAPIのアクセス経路の問題があり、そのままではWebhookは動きません。
* コントローラがAPIサーバにアクセスするときの権限が、クラスタ内で動作させた場合と異なります(例えば、`$HOME/.kube/config`の設定を利用すると、kubectlと同じ権限を持つようになります)。
* [Downward API](https://kubernetes.io/docs/tasks/inject-data-application/downward-api-volume-expose-pod-information/)が利用できません。

<!--
TODO: カスタムコントローラをTelepresenceで実行したときにキャッシュ周りの問題がありそう？

## Telepresenceによる実行

`make run`による実行では上記のようにいくつかの問題があります。
特にWebhookを手軽に動作させることができない点が不便です。

そこで、Telepresenceというツールを利用してコントローラをローカルのプロセスとして実行する方法を紹介します。

まずは下記のページを参考にしてTelepresence v2をインストールしてください。

* [Install Telepresence](https://www.telepresence.io/docs/latest/install/)

カスタムコントローラをTelepresenceで実行する場合、いくつか注意点があります。

Telepresenceでは、対象のワークロードにtraffic-agentというコンテナがインジェクトされるのですが、このコンテナはルート権限で実行する必要があります。
Kubebuilderが生成した[manager.yaml](../../codes/markdown-view/config/manager/manager.yaml)には、
SecurityContextで`runAsNonRoot: true`が指定されているので、これをコメントアウトする必要があります。

```yaml
      securityContext:
        runAsNonRoot: true
```

またTelepresenceでは、コンテナにマウントされたConfigMapやSecretが、ローカルのディレクトリにマウントされます。
そのため、ConfigMapやSecretにアクセスする際のパスが、Podとして実行する場合とTelepresenceで実行する場合で異なります。

ローカルにマウントされたディレクトリのパスは、環境変数`TELEPRESENCE_ROOT`で取得することができます。
Kubebuilderで生成したカスタムコントローラでは、Webhookの証明書のパスを下記のように設定し、`NewManager`するときのOptionとして指定しましょう。

[import:"telepresence,new-manager",unindent="true"](../../codes/markdown-view/main.go)

```go
	//! [telepresence]
	certDir := filepath.Join("tmp", "k8s-webhook-server", "serving-certs")
	root := os.Getenv("TELEPRESENCE_ROOT")
	fmt.Printf("TELEPRESENCE_ROOT: %s\n", root)
	time.Sleep(30 * time.Second)
	if len(root) != 0 {
		certDir = filepath.Join(root, certDir)
	} else {
		certDir = filepath.Join("/", certDir)
	}
	//! [telepresence]

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                 scheme,
		MetricsBindAddress:     metricsAddr,
		Port:                   9443,
		HealthProbeBindAddress: probeAddr,
		LeaderElection:         enableLeaderElection,
		LeaderElectionID:       "c124e721.zoetrope.github.io",
		CertDir:                certDir,
	})
```

なお、Telepresenceではローカルにボリュームをマウントするためにsshfsを利用しています。
ボリュームマウント機能がうまく動作しない場合は、sshfsがインストールされているかどうかを確認したり、
`/etc/fuse.conf`に下記のオプションが指定されているかどうかを確認してみてください。

```text
user_allow_other
```

準備が整ったら、[Kindで動かしてみよう](./kind.md)の手順通りにコントローラをデプロイします。

最後に下記のコマンドで、Kubernetes上で動いているコントローラを、make runを実行して起動したプロセスと置き換えます。

```console
telepresence intercept markdown-view-controller-manager --namespace markdown-view-system --service markdown-view-webhook-service --port=9443 -- make run
```

-->
