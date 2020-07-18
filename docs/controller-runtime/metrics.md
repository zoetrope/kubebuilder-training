# メトリクスの収集

## metricsListener

[import:"new-manager",unindent:"true"](../../codes/tenant/main.go)

これだけで、CPUやメモリの使用量などの基本的なメトリクスと、Reconcileにかかった時間やKubernetesクライアントのレイテンシーなど、controller-runtime関連のメトリクスが収集できるようになります。

## カスタムメトリクス

コントローラ固有のメトリクスを収集したい場合は、以下のようなコードを記述します。
詳しくは[Prometheusのドキュメント](https://prometheus.io/docs/instrumenting/writing_exporters/)を参照してください。

[import, title="metrics.go"](../../codes/tenant/controllers/metrics.go)

Reconcile処理の中で以下の処理を呼び出します。


[import:"create,metrics",unindent="true"](../../codes/tenant/controllers/tenant_controller.go)

## kube-rbac-proxy

[import:"bases,enable-prometheus,enable-auth-proxy"](../../codes/tenant/config/default/kustomization.yaml)

[import](../../codes/tenant/config/rbac/kustomization.yaml)

[import](../../codes/tenant/config/prometheus/monitor.yaml)



## Grafanaでの可視化

それでは実際にPrometheusとGrafanaを使って、コントローラのメトリクスを可視化してみましょう。

まずは[前章で解説した手順](../kubebuilder/kind.md)に従って、kind環境上にコントローラのデプロイをおこなってください。

Prometheus Operatorをセットアップするために、下記の手順に従ってHelmをインストールします。
- https://helm.sh/docs/intro/install/

つぎにHelmのリポジトリの登録をおこないます。

```
helm repo add stable https://kubernetes-charts.storage.googleapis.com/
```

Prometheus Operatorをセットアップします。完了まで少し時間がかかるので待ちましょう。

```
kubectl create ns prometheus
helm install prometheus stable/prometheus-operator --namespace=prometheus --set prometheus.prometheusSpec.serviceMonitorSelectorNilUsesHelmValues=false
kubectl wait pod --all -n prometheus --for condition=Ready --timeout 180s
```

起動したPrometheusにメトリクスを読み取る権限を付与する必要があるので、下記のマニフェストを適用します。

[import](../../codes/tenant/config/rbac/prometheus_role_binding.yaml)

```
kubectl apply -f ./config/rbac/prometheus_role_binding.yaml
```

ローカル環境からGrafanaの画面を確認するためにポートフォワードの設定をおこないます。

```
kubectl port-forward service/prometheus-grafana 3000:80 --address 0.0.0.0 --namespace prometheus
```

ブラウザから`http://localhost:3000`を開くとGrafanaの画面が確認できると思いますので、ユーザー名とパスワードを入力してログインします。
- Username: `admin`
- Password: `prom-operator`



```
histogram_quantile(0.99, sum(rate(controller_runtime_reconcile_time_seconds_bucket[5m])) by (le))
```
![grafana](./grafana.png)
