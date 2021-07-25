# Server Side Apply

UpdateやCreateOrUpdateは、GetしてからUpdateするまでの間に、他の誰かがリソースを書き換えてしまう可能性がある。すると情報が失われてしまう可能性。

Server-Side Apply方式では、リソースの各フィールドごとに管理者を記録することにより、複数のコントローラやユーザーが同一のリソースを編集した場合に衝突を検知することが可能です。
MergePatch方式ではそのような衝突検知はおこなわれません。

SSAでは、誰がどのフィールドを変更したのかを管理。
`--show-managed-fields`
もちろん、複数のコントローラが同じフィールドを別の値に書き換えようとした場合は、コンフリクトエラーとなります。

> オブジェクトのフィールドへの変更は、「フィールド管理」メカニズムによって追跡されます。フィールドの値が変更されると、所有権は現在の管理者から変更を行った管理者に移ります。オブジェクトを適用しようとしたときに、別のマネージャーが所有する異なる値のフィールドがあると、コンフリクトが発生します。これは、操作によって他の協力者の変更が取り消される可能性があることを知らせるために行われます。コンフリクトは強制的に発生させることができ、その場合は値が上書きされ、所有権が移されます。

ここではServer-Side Apply方式による`Patch()`の利用方法を紹介します。
以下の例では、Deploymentリソースの`spec.replicas`フィールドのみを更新しています。

なお、operation UpdateとApplyは別物。SSAで適用した場合operationはApplyになる。それ以外のCreateとかUpdateした場合はすべてUpdate。
これが異なると、FieldManagerが一致したとしても違うプログラムから更新されたものと見なされる。
なので、SSAでリソースを更新する場合は、CreateやUpdateを利用せずに、一貫してPatchを使いましょう。

YAMLから読み込む方式か、1個ずつ指定する方式か。

[import:"patch-apply"](codes/client-sample/main.go)

Server-Side Applyを利用するには、第3引数に`client.Apply`を指定し、オプションには`FieldManager`を指定する必要があります。
この`FieldManager`がフィールドごとの管理者の名前になるので、他のコントローラと被らないようにユニークな名前にしましょう。

なお、リストやマップをどのようにマージするのかは、Goの構造体に付与したマーカーで制御することが可能です。
詳しくは[Merge strategy](https://kubernetes.io/docs/reference/using-api/api-concepts/#merge-strategy)を参照してください。(TODO: あとで書く)

https://kubernetes.io/docs/reference/using-api/server-side-apply/#merge-strategy

全く同じ内容の要素をを別のコントローラが追加した場合どうなるか。多分追加できてしまう。そこでkeyを指定。markdownsもそうなのでは？やってみよう。

例えば[ServiceSpec](https://pkg.go.dev/k8s.io/api/core/v1#ServiceSpec)を見てみよう。

```yaml
ports:
- containerPort: 1053
  name: dns
  protocol: UDP
- containerPort: 1053
  name: dns-tcp
  protocol: TCP
```

```go
	// The list of ports that are exposed by this service.
	// More info: https://kubernetes.io/docs/concepts/services-networking/service/#virtual-ips-and-service-proxies
	// +patchMergeKey=port
	// +patchStrategy=merge
	// +listType=map
	// +listMapKey=port
	// +listMapKey=protocol
	Ports []ServicePort `json:"ports,omitempty" patchStrategy:"merge" patchMergeKey:"port" protobuf:"bytes,1,rep,name=ports"`
```

appsv1.Deploymentは使えない。なぜなら、書き換えたいフィールだけを明示する必要があるから。そのため型チェックがおこなえない問題点。
Kubernetes 1.21以降では、ApplyConfigurationという仕組みが用意され、
これは、すべてのフィールドがポインタ型で定義されている。nilに設定したフィールドは書き換えられないことを明示することができる。



CreateOrUpdateでDeploymentを作成した直後に、api-serverからそのDeploymentを取得して差分をチェックしてみましょう。
以下のような差分が生じます。

```diff
 spec:
+  progressDeadlineSeconds: 600
   replicas: 1
+  revisionHistoryLimit: 10
   selector:
     matchLabels:
       app.kubernetes.io/created-by: markdown-view-controller
       app.kubernetes.io/instance: markdownview-sample
       app.kubernetes.io/name: mdbook
+  strategy:
+    rollingUpdate:
+      maxSurge: 25%
+      maxUnavailable: 25%
+    type: RollingUpdate
   template:
     metadata:
+      creationTimestamp: null
       labels:
         app.kubernetes.io/created-by: markdown-view-controller
         app.kubernetes.io/instance: markdownview-sample
         app.kubernetes.io/name: mdbook
     spec:
       containers:
       - args:
         - serve
         - --hostname
         - 0.0.0.0
         command:
         - mdbook
         image: peaceiris/mdbook:latest
         imagePullPolicy: IfNotPresent
         livenessProbe:
+          failureThreshold: 3
           httpGet:
             path: /
             port: http
             scheme: HTTP
+          periodSeconds: 10
+          successThreshold: 1
+          timeoutSeconds: 1
         name: mdbook
         ports:
         - containerPort: 3000
           name: http
           protocol: TCP
         readinessProbe:
+          failureThreshold: 3
           httpGet:
             path: /
             port: http
             scheme: HTTP
+          periodSeconds: 10
+          successThreshold: 1
+          timeoutSeconds: 1
+        resources: {}
+        terminationMessagePath: /dev/termination-log
+        terminationMessagePolicy: File
         volumeMounts:
         - mountPath: /book/src
           name: markdowns
+      dnsPolicy: ClusterFirst
+      restartPolicy: Always
+      schedulerName: default-scheduler
+      securityContext: {}
+      terminationGracePeriodSeconds: 30
       volumes:
       - configMap:
+          defaultMode: 420
           name: markdowns-markdownview-sample
         name: markdowns
```

api-serverがデフォルト値を埋めたり、
また、それ以外にも何らかのMutating Webhookにより値が設定されたり、別のカスタムコントローラが値を書き換える場合もあります。
(例えばArgoCDでは、管理対象のリソースにラベルを付与する)

このようなことを考慮して、自分が書き換えたいフィールドだけを適切に設定することは難しい。

そこで、Server Side Apply
