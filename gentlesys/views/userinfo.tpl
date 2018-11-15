<!DOCTYPE html>
<html>
<head>
   <meta charset="utf-8">
    <meta name="viewport" content="width=device-width,initial-scale=1.0, minimum-scale=1.0, maximum-scale=1.0, user-scalable=no"/>
   <title>Gentlesys {{.Title}}</title>
   <link rel="stylesheet" href="http://apps.bdimg.com/libs/bootstrap/3.2.0/css/bootstrap.min.css">
   <script src="http://apps.bdimg.com/libs/jquery/2.1.4/jquery.min.js"></script>
   <script src="http://apps.bdimg.com/libs/bootstrap/3.2.0/js/bootstrap.min.js"></script>

   <style>
	.key-prob {
		color:#009900;
		font-size:20px;
		padding-right:20px;
	}
	.edit_prob {
		color:#009900;
	}
	.des-center {
	    text-align: center;
	    font-size:20px;
	    color:#4876FF;
	    padding-bottom:5px;
	}
	</style>
</head>
<body style="padding-top:55px;">
{{str2html .Navigation}}
<div class="container-fluid">

<div class="row">
    <div class="col-md-11 col-md-offset-2">
        <p class="key-prob">您的帖子详情(按照发布时间先后排名)</p>
        <ul class="pagination pagination-sm">
            <li><a class="btn btn-default" role="button">帖子索引</a></li>
        	<li><a href="{{.PrePage}}">&laquo;</a></li>
        	 {{range .RecordIndexs}}
            	<li class="{{.IsActive}}"><a href="{{.Ref}}">{{.Title}}</a></li>
        	{{end}}
        	<li><a href="{{.NextPage}}">&raquo;</a></li>
        </ul>
        
        <p>{{.Info}}</p>
        {{range .TopicsList}}
        <p>发布时间{{.Time}}&nbsp&nbsp[主题]&nbsp&nbsp<a href="/browse?sid={{.Sid}}&aid={{.Aid}}" target="_blank"> {{.Title}}</a>&nbsp&nbsp&nbsp&nbsp<a class="edit_prob" href="/edit?sid={{.Sid}}&aid={{.Aid}}" target="_blank">(编辑) </a></p>
        {{end}}
    </div>
</div>

</div>
</body>
</html>


