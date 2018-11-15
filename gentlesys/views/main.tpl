<!DOCTYPE html>
<html>
<head>
   <meta charset="utf-8">
   <meta name="viewport" content="width=device-width,initial-scale=1.0, minimum-scale=1.0, maximum-scale=1.0, user-scalable=no"/>
   <meta name="description" content="{{.Title}}"/>
   <title>{{.Title}} Powered by Gentlesys!</title>
   <link rel="stylesheet" href="http://apps.bdimg.com/libs/bootstrap/3.2.0/css/bootstrap.min.css">
   <script src="http://apps.bdimg.com/libs/jquery/2.1.4/jquery.min.js"></script>
   <script src="http://apps.bdimg.com/libs/bootstrap/3.2.0/js/bootstrap.min.js"></script>
   <style>
	.body-css {
		padding-top:55px;
	}
	.root-css {
	 background-color:#fcfdf8;
	}
   </style>
</head>

<body class="body-css">
{{str2html .Navigation}}

<div class="container-fluid">

	<div class="row">
	<div class="row root-css col-lg-10 col-lg-offset-1">
		<ul class="nav nav-pills" style="font-weight:bold;">
		{{range .Pagenav}}
            <li><a href="{{.Href}}">{{.Name}}</a></li>
        {{end}}
		</ul>
		
		<hr>
		
		{{range .Subject}}
		<div class="media col-lg-6">
	    <div class="pull-left">
	      <img src="/static/img/t1.png" class="media-object">
	    </div>
	    <div class="media-body">
	      <h5 class="media-heading"><a href="{{.Href}}">{{.Name}}</a></h5>
		  <small>帖数: {{.CurTopicIndex}}</small>
	      <small>{{.Desc}}</small>
	    </div>
		<hr>
	    </div>
		{{end}}
			
	</div>
	</div>
	
	

		
</div>
</body>

</html>

