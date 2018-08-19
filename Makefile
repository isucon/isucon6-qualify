.PHONY: main sysctl mysql nginx ctrl bench

main: sysctl mysql nginx ctrl link_init bench
	echo "OK"

sysctl:
	sudo ln -sf $(PWD)/sysctl.conf /etc/sysctl.conf
	sudo sysctl -a

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

link_init:
	sudo su - isucon -c 'ln -sf $(PWD)/init.sh init.sh'

bench:
	cd /home/isucon/gocode/src/github.com/isucon/isucon6-qualify/bench && ./bench --datadir=data -target=http://localhost
