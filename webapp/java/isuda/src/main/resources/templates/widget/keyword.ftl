<article>
  <h1><a href="/keyword/${entry.keyword?url('UTF-8')}">${entry.keyword}</a></h1>
  <div>${entry.html}</div>
  <button class="js-add-star" data-keyword="${entry.keyword}" data-user-name="${userName}"><img src="/img/star.gif"></button>
  <span class="js-stars" data-keyword="${entry.keyword}">
<#list entry.stars as star>
      <img src="/img/star.gif" title="${star.user_name}" alt="${star.user_name}">
</#list>
  </span>
</article>