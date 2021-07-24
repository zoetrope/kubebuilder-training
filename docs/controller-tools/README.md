# controller-tools

Kubebuilderでは、カスタムコントローラの開発を補助するためのツール群として[controller-tools](https://github.com/kubernetes-sigs/controller-tools)を提供しています。

controller-toolsには下記のツールが含まれていますが、本資料ではcontroller-genのみを取り扱います。

- controller-gen
- type-scaffold
- helpgen

##  controller-gen

`controller-gen`は、GoのソースコードをもとにしてマニフェストやGoのソースコードの生成をおこなうツールです。

`controller-gen`のヘルプを確認すると、下記の5種類のジェネレータが存在することがわかります。

```
generators

+webhook                                                                                                  package  generates (partial) {Mutating,Validating}WebhookConfiguration objects.
+schemapatch:manifests=<string>[,maxDescLen=<int>]                                                        package  patches existing CRDs with new schemata.
+rbac:roleName=<string>                                                                                   package  generates ClusterRole objects.
+object[:headerFile=<string>][,year=<string>]                                                             package  generates code containing DeepCopy, DeepCopyInto, and DeepCopyObject method implementations.
+crd[:crdVersions=<[]string>][,maxDescLen=<int>][,preserveUnknownFields=<bool>][,trivialVersions=<bool>]  package  generates CustomResourceDefinition objects.
```

`kubebuilder`が生成したMakefileには、`make manifests`と`make generate`というターゲットが用意されており、`make manifests`では`webhook`, `rbac`, `crd`の生成、`make generate`では`object`の生成がおこなわれます。

`controller-gen`がマニフェストの生成をおこなう際には、Goのstructの構成と、ソースコード中に埋め込まれた`// +kubebuilder:`から始まるコメント(マーカーと呼ばれる)を目印にします。

利用可能なマーカーは下記のコマンドで確認することができます。(`-ww`や`-www`を指定するとより詳細な説明が確認できます)

```console
$ controller-gen crd -w
$ controller-gen webhook -w
```
