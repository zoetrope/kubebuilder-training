# Reconcile

## Reconcileとは

## Reconcileはいつ呼ばれるのか

* コントローラが扱うリソースが更新されたとき
* コントローラの起動時
* 外部イベント
* Resync

[import:"pred,managedby",unindent:"true"](../../codes/tenant/controllers/tenant_controller.go)

## Reconcileの実装

[import:"reconcile"](../../codes/tenant/controllers/tenant_controller.go)
