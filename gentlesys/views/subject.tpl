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
	<div class="row  col-lg-10 col-lg-offset-1">
      
	  <div class="root-css col-lg-8">
	  
	  <ul class="breadcrumb">
       <li><a href="/">首页</a></li>
       <li class="active"><a href="{{.HrefSub}}">{{.SubName}}</a></li>
       </ul>
       
	  <ul class="pagination pagination-sm">
        <li><a class="btn btn-default" href="/article{{.Args}}" role="button">发帖</a></li>
    	<li><a href="{{.PrePage}}">&laquo;</a></li>
    	 {{range .RecordIndexs}}
        	<li class="{{.IsActive}}"><a href="{{.Ref}}">{{.Title}}</a></li>
    	{{end}}
    	<li><a href="{{.NextPage}}">&raquo;</a></li>
      </ul>
		
	  <div class="list-group">
	  {{range .Topic}}
    	  <li class="list-group-item">
    	  <a href="{{.Href}}">{{.Name}}</a>
    	  <span class="badge">热点</span>
    	  <p class="list-group-item-text"><small>[ArName]&nbsp;&nbsp;•&nbsp;[Data]&nbsp;发布&nbsp;&nbsp;•&nbsp;[200000/20000000]</small></p>
    	  
    	  </li>
	  {{end}}
	  </div>
	  </div>
			
	</div>
	</div>
</div>
</body>

</html>

