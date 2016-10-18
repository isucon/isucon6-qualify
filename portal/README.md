# isucon6q-portal

ISUCON6 予選ポータルサイトです。

## 運営アカウント

どの日も同じアカウントで入れます。

- ID: 9999
- PASS: `eimae5eebocheim4Kool`

他のチームと同様の扱いでログインできます。

## デプロイ

~~~
Host isucon6q-portal
    User isucon
    HostName 13.78.94.217
~~~

で

    make deploy TARGET=isucon6q-portal ANSIBLE_ARGS=-vv

## 起動オプション

- `-database=dsn <dsn="root:@/isu6qportal">`
- `-starts-at <hour=10>`
- `-ends-at <hour=18>`

## 運用

終了一時間前あたりで `team_scores_snapshot` テーブルを作るとリーダーボードが固定されます。

    INSERT INTO team_scores_snapshot SELECT * FROM team_scores

## 開発・運用むけ情報

秘密のURLです。認証とかはとくになし

- /top4aew4fe9yeehu/debug/vars
- /top4aew4fe9yeehu/debug/queue
- /top4aew4fe9yeehu/debug/leaderboard
