<#include "/templates/header.ftl">

<form class="form" action="/keyword" method="POST">
  <ul>
    <li><input type="text" name="keyword"></li>
    <li><textarea name="description"></textarea></li>
    <li><input class="btn btn-primary" type="submit" value="Post" /></li>
  </ul>
</form>

<#list entries as entry>

  <#include "/templates/widget/keyword.ftl">

</#list>

<nav class="pagination">
<#if 1 < page>
  <li><a href="?page=${page - 1}">&laquo;</a></li>
<#else>
  <li class="disabled"><span>&laquo;</span></li>
</#if>
<#list pages as p>
  <li<#if p == page> class="active"</#if>><a href="?page=${p}">${p}</a></li>
</#list>
<#if page == lastPage>
  <li class="disabled"><span>&raquo;</span></li>
<#else>
  <li><a href="?page=${page + 1}">&raquo;</a></li>
</#if>
  </ul>
</nav>

<#include "/templates/footer.ftl">