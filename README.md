# ISUCON6 予選リポジトリ

レギュレーションは Regulation.md に書かれています。

## リポジトリ内容

- `bench/` ベンチマーカー及びワーカー
- `cmd/isupam/` スパムチェッカーisupamのソースコード
- `cmd/importpages/` Wikipediaデータインポーター
- `db/` 競技用のMySQL初期データのSQL
- `portal/` 競技用ポータルのソースコード
- `provisioning/` Asure deployに利用したansible playbook
- `tools/` 問題作成時に利用したスクリプト群
- `webapp/` 予選用各言語参考実装ファイル等

## 利用したデータについて

### キーワードデータ

[日本語版Wikipediaの2016年7月1日時点のdumpデータ](https://dumps.wikimedia.org/jawiki/20160701/jawiki-20160701-pages-articles-multistream.xml.bz2)を利用しています。

日本語版Wikipediaのライセンスに準じ、以下の文書素材は CC-BY-SA 3.0(<https://creativecommons.org/licenses/by-sa/3.0/>)に従って公開されています。

- db/isuda_entry.sql
- bench/isucon6q/data/ng.json
- bench/isucon6q/data/ok.json
- bench/isucon6q/data/init.json
- bench/isucon6q/data/year.json

これらのファイルは、  cmd/importpages/ と tools/ 以下のスクリプトファイルを用いて生成されました。利用されているデータの二次利用元のURLは URLs.txt に記載されています。

### スパムワード一覧

以下のサイトの放送禁止用語一覧を利用しました。公開にあたりサイト管理者の許諾を得ています。

http://monoroch.net/kinshi/  
http://monoroch.net/kinshi/kinshi.csv
