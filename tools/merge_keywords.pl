#!/usr/bin/env perl
use 5.014;
use warnings;
use utf8;
use autodie;
use lib 'webapp/perl/lib';

=head1 DESCRIPTION

initj.jsonとyear.jsonにリンク情報とlength情報を付加する

=cut

use Path::Tiny qw/path/;
use JSON::XS qw/decode_json/;

my $rel_file = '.tmp/keyword_rel.json';

my @keywords = do {path($rel_file)->lines};
my %rel_hash;
for my $line (@keywords) {
    my $data = decode_json $line;
    $rel_hash{$data->{k}} = $data;
}

my $json_driver = JSON::XS->new->utf8(1)->canonical(1);
sub {
    my $f = shift;
    my @keywords = do {path($f)->lines};
    eval {unlink $f};
    open my $fh, '>>', $f;
    for my $l (@keywords) {
        my $d = decode_json $l;
        my $rel = $rel_hash{$d->{k}};
        $d->{links} = $rel->{links};
        $d->{length} = $rel->{length};
        print $fh $json_driver->encode($d) . "\n";
    }
}->($_) for qw{bench/data/init.json bench/data/year.json};
