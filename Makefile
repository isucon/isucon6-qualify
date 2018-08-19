.PHONY: main sysctl mysql nginx ctrl bench

mysql: 
	sudo systemctl restart mysql

nginx: 
	sudo rm -f /var/log/nginx/access.log /var/log/nginx/error.log
	sudo touch /var/log/nginx/access.log /var/log/nginx/error.log
	sudo systemctl reload nginx

app:
	cd /home/isucon/isucon6-qualify/webapp/go && make
	sudo systemctl restart isutar.go
	sudo systemctl restart isuda.go

bench:
	cd /home/isucon/gocode/src/github.com/isucon/isucon6-qualify/bench && ./bench --datadir=data -target=http://localhost
