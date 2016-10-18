#!/usr/bin/env perl
use 5.014;
use warnings;
use utf8;
use autodie;
use lib 'webapp/perl/lib';

=head1 DESCRIPTION

データを投入する

=cut

use Path::Tiny qw/path/;
use JSON::XS qw/decode_json/;

use Isuda::Web;

my $limit = shift;
$limit ||= 6000;

my $WORK_DIR = $ENV{ISUDA_DATA_DIR} || '.tmp';
my $year_file = "$WORK_DIR/wikipedia_year.json";
my $file = "$WORK_DIR/wikipedia_ok.json";

my @keywords = path($file)->lines;
if (scalar(@keywords) > $limit) {
    @keywords = splice @keywords, 0, $limit;
}
unshift @keywords, path($year_file)->lines;

my @vars;
for my $k (@keywords) {
    my $d = decode_json $k;
    push @vars, $d->{k}, $d->{v};
}

my $dbh = Isuda::Web->new->dbh;
$dbh->do('TRUNCATE entry');
while (my @v = splice @vars, 0, 1000) {
    my $sql = 'INSERT INTO entry (author_id, keyword, description, created_at, updated_at) VALUES ' .
        ('(1, ?, ?, NOW(), NOW()),' x (scalar(@v)/2));
    chop $sql; # cut off last comma
    $dbh->query($sql, @v);
}
