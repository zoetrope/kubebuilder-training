# make runによるデバッグ

コントローラの開発中は、何度もビルドし直して
コンテナイメージのビルドには時間がかかります。
また、コンテナとして動作しているプログラムはデバッガによるデバッグ実行がしにくいという問題もあります。

Kubebuilderによって生成されたMakefileには`make run`というターゲットが用意されており、コントローラをローカル環境で動作させることが可能です。

ただし、Kubernetesクラスタ上で動作させる場合といくつか違いがあるので、注意して利用してください。

* リーダーエレクション機能を利用する場合は、[Options](https://pkg.go.dev/sigs.k8s.io/controller-runtime/pkg/manager?tab=doc#Options)の`LeaderElectionNamespace`を指定する必要があります。
* Webhook機能を動かすためには下記の設定が必要になります。
  * Webhook API用の証明書を用意します(cert-managerが生成したSecretから読み出すなど)
  * [Options](https://pkg.go.dev/sigs.k8s.io/controller-runtime/pkg/manager?tab=doc#Options)の`CertDir`に証明書のパスを指定します
  * `MutatingWebhookConfiguration`, `ValidatingWebhookConfiguration`のマニフェストに、証明書の設定と、ローカルのWebhook APIのURLを指定します。
* コントローラがAPIサーバにアクセスするときの権限が、クラスタ内で動作させた場合と異なります(例えば、`$HOME/.kube/config`の設定を利用すると、kubectlと同じ権限を持つようになります)。
* [Downward API](https://kubernetes.io/docs/tasks/inject-data-application/downward-api-volume-expose-pod-information/)が利用できません。
