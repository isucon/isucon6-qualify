#!/usr/bin/env plackup
use 5.014;
use warnings;

use FindBin;
use lib "$FindBin::Bin/lib";
use File::Spec;
use Plack::Builder;

use Isutar::Web;

my $app = Isutar::Web->psgi($FindBin::Bin);
builder {
    enable 'ReverseProxy';
    $app;
};
