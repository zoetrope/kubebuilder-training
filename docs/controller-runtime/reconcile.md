# Reconcileの実装

## Reconcileとは

## Reconcileはいつ呼ばれるのか

* コントローラが扱うリソースが更新されたとき
* コントローラの起動時
* 外部イベント
* Resync

権限の確認
```console
kubectl get all -n test1 --as=system:serviceaccount:default:default
```
