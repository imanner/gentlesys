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
		
	  <hr>
      
	  <div class="root-css col-lg-8">
	  <ul class="breadcrumb">
       <li><a href="/">首页</a></li>
       <li><a href="{{.HrefSub}}">{{.SubName}}</a></li>
       </ul>
	  <div class="list-group">
	  {{range .Topic}}
    	  <li class="list-group-item">
    	  <a href="{{.Href}}">{{.Name}}</a>
    	  <span class="badge">14</span>
    	  <p class="list-group-item-text">这里是其他说明</p>
    	  </li>
	  {{end}}
	  </div>
	  </div>
			
	</div>
	</div>
</div>
</body>

</html>

