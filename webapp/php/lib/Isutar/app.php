<?php
namespace Isutar\Web;

use Slim\Http\Request;
use Slim\Http\Response;
use PDO;
use PDOWrapper;

$container = new class extends \Slim\Container {
    public $dbh;
    public function __construct() {
        parent::__construct();

        $this->dbh = new PDOWrapper(new PDO(
            $_ENV['ISUTAR_DSN'],
            $_ENV['ISUTAR_DB_USER'] ?? 'isucon',
            $_ENV['ISUTAR_DB_PASSWORD'] ?? 'isucon',
            [ PDO::MYSQL_ATTR_INIT_COMMAND => "SET NAMES utf8mb4" ]
        ));
    }
};
$app = new \Slim\App($container);

$app->get('/initialize', function (Request $req, Response $c) {
    $this->dbh->query('TRUNCATE star');
    return render_json($c, [
        'result' => 'ok',
    ]);
});

$app->get('/stars', function (Request $req, Response $c) {
    $stars = $this->dbh->select_all(
        'SELECT * FROM star WHERE keyword = ?'
    , $req->getParams()['keyword']);

    return render_json($c, [
        'stars' => $stars,
    ]);
});

$app->post('/stars', function (Request $req, Response $c) {
    $keyword = $req->getParams()['keyword'];

    $origin = $_ENV['ISUDA_ORIGIN'] ?? 'http://localhost:5000';
    $url = "$origin/keyword/" . rawurlencode($keyword);
    $ua = new \GuzzleHttp\Client;
    try {
        $res = $ua->request('GET', $url)->getBody();
    } catch (\Exception $e) {
        return $c->withStatus(404);
    }

    $this->dbh->query(
        'INSERT INTO star (keyword, user_name, created_at) VALUES (?, ?, NOW())',
        $keyword,
        $req->getParams()['user']
    );
    return render_json($c, [
        'result' => 'ok',
    ]);
});

$app->run();
