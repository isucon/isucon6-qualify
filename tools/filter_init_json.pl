#!/usr/bin/env perl
use 5.014;
use warnings;
use utf8;
use autodie;

use Path::Tiny qw/path/;
use JSON::XS qw/decode_json/;

my $f = 'bench/data/init.json';
my $json_driver = JSON::XS->new->utf8(1)->canonical(1);
my @keywords = do {path($f)->lines};
eval {unlink $f};
open my $fh, '>>', $f;
for my $l (@keywords) {
    my $d = decode_json $l;

    $d->{links} = [
        grep { $_ !~ /^(?:Java Platform, Enterprise Edition|M\@M)$/ }
        @{ $d->{links} }
    ];
    print $fh $json_driver->encode($d) . "\n";
}
