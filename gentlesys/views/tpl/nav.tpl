<nav class="navbar navbar-default navbar-fixed-top">
    <div class="container-fluid">
    <div class="row">
        <div class="col-md-offset-1">
		 <div class="navbar-header">
	     <button type="button" class="navbar-toggle collapsed" data-toggle="collapse" data-target="#navbar" aria-expanded="false" aria-controls="navbar">
         <span class="sr-only">Toggle navigation</span>
         <span class="icon-bar"></span>
         <span class="icon-bar"></span>
         <span class="icon-bar"></span>
         </button>
	  <a class="navbar-brand" style="color:#8470FF;padding-left:15px;text-decoration:none" href="/">{{.NavHead}}</a>
	</div>
	<div id="navbar" class="navbar-collapse collapse">
    <ul class="nav navbar-nav">
	 {{range .NavNodes}}
	  <li><a href="{{.Href}}">{{.Name}}</a></li> 
	{{end}}
    </ul>
	 </div>
    </div>
    </div>
    </div>
</nav>