package main

import (
	"sync"

	"github.com/isucon/isucon6-qualify/bench/checker"
)

// 普通のページに表示されるべき静的ファイルに一通りアクセス
func loadAssets(s *checker.Session) {
	actions := []*checker.AssetAction{
		checker.NewAssetAction("/favicon.ico", &checker.Asset{MD5: "07b21a6c8984e04d108064c585411601"}),
		checker.NewAssetAction("/css/bootstrap.min.css", &checker.Asset{MD5: "4082271c7f87b09c7701ffe554e61edd"}),
		checker.NewAssetAction("/css/bootstrap.min.css", &checker.Asset{MD5: "4082271c7f87b09c7701ffe554e61edd"}),
		checker.NewAssetAction("/css/bootstrap-responsive.min.css", &checker.Asset{MD5: "f889adb0886162aa4ceab5ff6338d888"}),
		checker.NewAssetAction("js/jquery.min.js", &checker.Asset{MD5: "05e51b1db558320f1939f9789ccf5c8f"}),
		checker.NewAssetAction("js/bootstrap.min.js", &checker.Asset{MD5: "d700a93337122b390b90bbfe21e64f71"}),
		checker.NewAssetAction("img/star.gif", &checker.Asset{MD5: "b9492ba28fe93e6ea0a3d705ca0cbfde"}),
	}

	var wg sync.WaitGroup
	for _, a := range actions {
		wg.Add(1)
		go func(a *checker.AssetAction) {
			defer wg.Done()
			a.Play(s)
		}(a)
	}
	wg.Wait()
}
