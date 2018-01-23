# ISUCON6 予選 Javaプログラムリポジトリ

## 利用しているもの

- マイクロフレームワーク [Spark](http://sparkjava.com/)
- テンプレートエンジン [freemarker](https://freemarker.apache.org/)


## 利用の前提条件

Java以外の環境が構築済みであること。

https://github.com/matsuu/vagrant-isucon/tree/master/isucon6-qualifier


## 起動方法

/home/isucon/webapp に、このディレクトリの内容をすべてコピーし、deploy.sh を実行してください。

その後、他言語と同様でsystemdを利用して起動、停止できます。

```sh
cd /home/isucon/webapp/java
chmod 755 deploy.sh build.sh
./deploy.sh

systemctl start isuda.java
systemctl start isutar.java
```

## ビルド方法

build.sh を実行してください。

```sh
./build.sh
```
