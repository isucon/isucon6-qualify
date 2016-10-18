$('.js-add-star').on('click', function() {
    var elem = this;
    var keyword = elem.getAttribute('data-keyword');
    var userName = elem.getAttribute('data-user-name');
    if (!keyword || !userName) {
        alert('Please login.');
        return
    }
    $.post('/stars', {
        keyword: keyword,
        user: userName
    }).done(function() {
        $('.js-stars').filter(function() {
            return this.getAttribute('data-keyword') == keyword;
        }).trigger('addStar');
    });
});

$('.js-stars').on('addStar', function() {
    var elem = this;
    var star = document.createElement('img');
    star.setAttribute('src', '/img/star.gif')
    elem.appendChild(star);
});
