#!/usr/bin/env perl
use 5.014;
use warnings;
use utf8;
use autodie;
use lib 'lib';

use Acme::CPANAuthors::Japanese;
my @users = map {lc} sort keys %{Acme::CPANAuthors::Japanese->authors};

use Isuda::Web;

my $dbh = Isuda::Web->new->dbh;
$dbh->do('TRUNCATE user');

for my $u (@users) {
    Isuda::Web::register($dbh, $u, $u);
}
