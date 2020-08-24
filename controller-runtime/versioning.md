# CRDのバージョニング

バージョニング難しい。

例えば、CRDを後方互換性のない形で変更してしまい、
カスタムコントローラの利用ユーザーが、カスタムリソースを一旦削除しなければならない。
サービスの停止、
非常に手間がかかる、
コントローラの種類によってはデータが失われてしまうようなケースも発生します。

後方互換性のある形で変更しなければなりません。

本資料のなかでこれまでつくってきたカスタムリソースは以下のようなものでした。

[import](../../codes/tenant/config/samples/multitenancy_v1_tenant.yaml)

フィールドの追加や

[import](../../codes/tenant/config/samples/multitenancy_v1_1_tenant.yaml)

adminフィールドは現在ひとつの値しか指定できませんが、これを複数指定できるように

[import](../../codes/tenant/config/samples/multitenancy_v1_2_tenant.yaml)

adminとadminsフィールドが存在するのはユーザーにとっては利用しにくいものです。
そこで、下記のようにadminsフィールドにまとめたい。
この場合は互換性がなくなってしまうので、apiVersionをv2にして

[import](../../codes/tenant/config/samples/multitenancy_v2_tenant.yaml)

そしてconversion webhookを用意します。
