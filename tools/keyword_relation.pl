#!/usr/bin/env perl
use 5.014;
use warnings;
use utf8;
use autodie;
use lib 'webapp/perl/lib';

=head1 DESCRIPTION

キーワード同士の親子関係を出力する

=cut

use List::MoreUtils qw/uniq/;
use Encode qw/encode_utf8/;
use JSON::XS qw/encode_json/;

open my $fh, '>>', '.tmp/keyword_rel.json';

use Isuda::Web;
my $dbh = Isuda::Web->new->dbh;

my $keywords = $dbh->select_all(qq[
    SELECT * FROM entry ORDER BY CHARACTER_LENGTH(keyword) DESC
]);
my $re = join '|', map { quotemeta $_->{keyword} } @$keywords;
   $re = "($re)";

my $json_driver = JSON::XS->new->utf8(1)->canonical(1);

my $sth = $dbh->prepare('SELECT keyword, description FROM entry ORDER BY id');
$sth->execute;
while (my $row = $sth->fetchrow_hashref) {
    my $content = $row->{description};
    my @links;
    while ($content =~ /$re/g) {
        push @links, $1;
    }
    @links = uniq sort @links;
    my $length = length(encode_utf8($content));
    say $fh $json_driver->encode({
        k      => $row->{keyword},
        links  => \@links,
        length => $length,
    });
}
