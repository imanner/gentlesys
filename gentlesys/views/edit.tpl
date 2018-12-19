<!DOCTYPE html>
<html>
<head>
   <meta charset="utf-8">
    <meta name="viewport" content="width=device-width,initial-scale=1.0, minimum-scale=1.0, maximum-scale=1.0, user-scalable=no"/>
   <title>Gentlesys {{.Title}}</title>
   <link rel="stylesheet" href="http://apps.bdimg.com/libs/bootstrap/3.2.0/css/bootstrap.min.css">
   <script src="http://apps.bdimg.com/libs/jquery/2.1.4/jquery.min.js"></script>
   <script src="http://apps.bdimg.com/libs/bootstrap/3.2.0/js/bootstrap.min.js"></script>
   <script src="/static/bootsp/edt/kindeditor-all-min.js"></script>
   <script src="/static/bootsp/edt/lang/zh-CN.js"></script>
   <style>
    h3,h5 {
   	color:#009900;
   	}
	.key-prob {
		color:#000066;
		padding-right:20px;
	}
	</style>
</head>
<body style="padding-top:55px;">
{{str2html .Navigation}}
<div class="container-fluid">
<div class="row">
<div class="col-md-8 col-md-offset-2">
<form role="form">
    <div class="form-group">
        <h3>您正在编辑帖子</h3>
        <p><span class="key-prob">作者信息</span> {{.UserName}}     <label class="btn">
		<input id="anonymity" type="checkbox" {{.Check}} autocomplete="off"> 匿名发表
        </label></p>
        <h5><strong>博文题目 (标题不能修改)</strong></h5>
    	<textarea id="title" class="form-control" rows="1" readonly="readonly"></textarea>
    	<script>document.getElementById("title").value="{{.Title}}"</script>
    	{{if .IsNotice}}
    	<p style="margin-top:10px;">
        <span><strong>请选择公告所在的版块</strong></span>
        <span><select id="tcType" class="select">
    	{{range .SubType}}
        <option value="{{.UniqueId}}">{{.Name}}</option>
        {{end}}
        </select>
        </span>
        </p>
        {{else}}
    	<p style="margin-top:10px;"><span>
    	<strong>请选择文章类型</strong>
    	</span>
    	<span><select id="tcType" class="select">
    	{{range $index, $elem := .TopicType}}
        <option value="{{$index}}">{{$elem}}</option>
        {{end}}
        </select>
        </span></p>
        {{end}}
        
    	<h5><strong>文章内容 (必填，提示：html字数不能超过1000000，如果字数较多，请先写好后一并粘贴过来，以防文字丢失！)</strong></h5>
        <div>
        <textarea id="story" name="content" style="width:100%;height:500px;"></textarea>
        <script>
        	$("#story").html({{.Story}})
        </script>
        </div>
    	<script type="text/javascript">

		$("#tcType").val({{.SelValue}}); 
		
		var ke
        KindEditor.ready(function(K) {
            ke = K.create('#story', {
            pasteType:1,
            allowImageUpload:true,
            allowFlashUpload:false,
            allowMediaUpload:false,
            allowFileUpload:false,
            cssData: 'body {font-family: "微软雅黑"; font-size: 18px;}',
            items:[  'source', '|', 'preview', 'code', '|', 'justifyleft', 'justifycenter', 'justifyright',
        'justifyfull','selectall', '|',
        'formatblock', 'fontsize', 'removeformat','|', 'forecolor', 'bold',
        'italic', 'underline', '|', 'image', 'media','link', '|'],
		afterCreate:function () {
            this.sync();                  
            },
            afterChange:function() {
            this.sync();
            }
        });
        });
        
    	function send(){
    		var any = document.getElementById("anonymity");
			var anonymity = false;
			if(any.checked)
			{
				anonymity = true;
			}
			
    		var title = document.getElementById("title").value;
			if (title.length > 128) {
        		document.getElementById("info").innerHTML=("分享题目 超过128字，请删减一些，目前有长度是" + title.length);
				return;
			} else if (title.length < 1) {
				document.getElementById("info").innerHTML=("分享题目 是空的，请填写");
				return;
			}
			
			
			var options=$("#tcType option:selected");
			var type = options.val(); 
			
			var count = ke.count()
            if(count < 1) {
				document.getElementById("info").innerHTML=("文章内容 是空的，请填写");
				return;
			}
			var story = ke.html();
			
			document.getElementById("info").innerHTML=("正在提交中......");
			
			var btn = $("#sendButton"); 
			btn.attr("disabled", true);

			$.ajax({
	          async:true,
	          cache:false,
	          timeout:10000,
	          type:"POST",
	          url:"/article",
	          data:{
	            atId_:{{.ArtiId}},
	          	subId_:{{.Sid}},
	          	userId_:{{.UserId}},
				userName_:{{.UserName}},
		        title_:title,
				type_:type,
				story_:story,
				anonymity_:anonymity
		    	},
	          error:function(jqXHR, textStatus, errorThrown){
	            if(textStatus=="timeout"){
	              document.getElementById("info").innerHTML=("提交超时，请重试...");
	            }else{
	              document.getElementById("info").innerHTML=("提交失败!");
	            }
				btn.attr("disabled", false);
	          },
	          success:function(msg){
	          	if ("[0]" != msg.substr(0,3)) {
					if ("[4]" != msg.substr(0,3)) {
						btn.attr("disabled", false);
					}
					document.getElementById("info").innerHTML=(msg);
	          	} else {
					window.location.href=msg.substr(3)
	          	}
	            
	          }
	        });

        }
    	</script>
    	<button id="sendButton" type="button" class="btn btn-default btn-sm" style = "float: right;" onclick="send()">提交分享</button>
        <p id="info" class="alert alert-info">点击右边的 提交分享 按钮进行提交！</p>
    </div>
</form>
</div>
</div>

</div>
</body>
</html>
