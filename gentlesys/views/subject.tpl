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
	<div class="col-lg-10 col-lg-offset-1">
      
	  <div class="root-css col-lg-8">
      <p class="crumbs"><a href="/">首页</a> &gt;  <a href="{{.HrefSub}}">{{.SubName}}</a>&gt;<a href="/article{{.Args}}">[我要发帖]</a></p>

	  <ul class="pagination pagination-sm">
        <li><a class="btn btn-default" href="/article{{.Args}}" role="button">发帖</a></li>
    	<li><a href="{{.PrePage}}">&laquo;</a></li>
    	 {{range .RecordIndexs}}
        	<li class="{{.IsActive}}"><a href="{{.Ref}}">{{.Title}}</a></li>
    	{{end}}
    	<li><a href="{{.NextPage}}">&raquo;</a></li>
      </ul>
	  <table class="table table-condensed mtable-css"> 
        <tbody>
         {{range .Topic}}
         <tr>
            <td>
            <h5> <a href="/browse?sid={{$.Sid}}&aid={{.Id}}">{{.Title}}</a></h5>
            <small>{{.UserName}}&nbsp;&nbsp;•&nbsp;{{.Date}}&nbsp;发布&nbsp;&nbsp;•&nbsp;[{{.ReadTimes}}/{{.ReplyTimes}}]</small>
            </td>
         </tr>
         {{end}}
         </tbody>
       </table>
       {{if .NoMore}}
        <hr />
        <p>没有更多帖子了...</p>
       {{end}}
	  </div>
			
	</div>
	</div>
</div>
</body>

</html>

