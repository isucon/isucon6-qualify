# ISUCON6予選用ベンチマーカー

1. workerはportalから定期的にジョブを取得する
2. 取得したジョブ情報を元に、workerはbenchmarkerをキックする
3. benchmarkerが標準出力に吐いたJSONを結果として、workerはportalに結果を投稿する

- ./
  - ベンチマーカー用のコード
- worker/
  - ワーカー
- isucon6q/
  - 実行ファイル類を直置きする
  - /home/isucon/isucon6q にそのまま配置される想定
