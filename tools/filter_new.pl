#!/usr/bin/env perl
use 5.014;
use warnings;
use utf8;
use autodie;
use lib 'webapp/perl/lib';

=head1 DESCRIPTION

リンクが新たに作られそうなキーワードを抽出する

=cut

use Path::Tiny qw/path/;
use JSON::XS qw/decode_json/;
use Encode qw/encode_utf8/;
use Isuda::Web;
use List::UtilsBy qw/rev_nsort_by/;

my $dbh = Isuda::Web->new->dbh;

# okのラスト30000行
my @lines = path('.tmp/wikipedia_ok.json')->lines;
   @lines = @lines[-30000..-1];

my @rows;
my $last = sub {
    @rows = rev_nsort_by { length(encode_utf8($_->{k})) } @rows;
    open my $fh, '>', 'bench/data/ok.json';
    my $json_driver = JSON::XS->new->utf8(1)->canonical(1);
    for my $row (@rows) {
        print $fh $json_driver->encode($row) . "\n";
    }
    exit 0;
};
$SIG{INT} = $last;

my $count = 0;
for my $line (@lines) {
    $count++;
    unless ($count % 100) {
        warn sprintf("done: %d, found %d\n", $count, scalar(@rows));
    }

    my $j = decode_json $line;

    my $quoted_word = $j->{k};
    $quoted_word =~ s/\\/\\\\/g;
    $quoted_word =~ s/'/\\'/g;
    $quoted_word = "%$quoted_word%";

    my $sql1 = 'SELECT keyword FROM entry WHERE keyword LIKE ?';
    my $contains = $dbh->select_all($sql1, $quoted_word);
    next if @$contains;

    my $sql2 = 'SELECT keyword FROM entry WHERE description LIKE ?';
    my @links = map { $_->{keyword} } @{ $dbh->select_all($sql2, $quoted_word) };
    next unless @links;
    warn encode_utf8(sprintf("%s: found %d links\n", $j->{k}, scalar(@links)));

    my $length = length(encode_utf8($j->{v}));
    push @rows, {
        k          => $j->{k},
        back_links => \@links,
        v          => $j->{v},
        length     => $length,
    };
}
$last->();
