<#include "/templates/header.ftl">

<h2>${action?capitalize}</h2>
<form class="form" action="/${action}" method="POST">
  ID: <input type="text" name="name" value="">
  PW: <input type="password" name="password" value="">
  <p><input class="btn btn-primary" type="submit" value="${action?capitalize}" /></p>
</form>


<#include "/templates/footer.ftl">