<?php
require __DIR__ . '/vendor/autoload.php';
require __DIR__ . '/lib/glue.php';

$_ENV['ISUDA_DSN'] = 'mysql:host=127.0.0.1;dbname=isuda';
$_ENV['PHP_TEMPLATE_PATH'] = __DIR__ . '/views';

$_SERVER += ['PATH_INFO' => $_SERVER['REQUEST_URI']];
$_SERVER['SCRIPT_NAME'] = '/' . basename($_SERVER['SCRIPT_FILENAME']);
$file = dirname(__DIR__) . '/public' . $_SERVER['REQUEST_URI'];
if (is_file($file)) {
    if (PHP_SAPI == 'cli-server') return false;
    $mimetype = [
        'js' => 'application/javascript',
        'css' => 'text/css',
        'ico' => 'image/vnd.microsoft.icon',
    ][pathinfo($file, PATHINFO_EXTENSION)] ?? false;
    if ($mimetype) {
        header("Content-Type: {$mimetype}");
        echo file_get_contents($file); exit;
    }
}

session_start();

require __DIR__ . '/lib/Isuda/app.php';
