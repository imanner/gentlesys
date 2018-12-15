<!DOCTYPE html>
<html>
<head>
	<meta charset="utf-8"> 
	<meta name="viewport" content="width=device-width,initial-scale=1.0, minimum-scale=1.0, maximum-scale=1.0, user-scalable=no"/>
   <title>Gentlesys {{.Title}}</title>
   <link rel="stylesheet" href="http://apps.bdimg.com/libs/bootstrap/3.2.0/css/bootstrap.min.css">
   <script src="http://apps.bdimg.com/libs/jquery/2.1.4/jquery.min.js"></script>
   <script src="/static/bootsp/js/jquery.md5.js"></script>
    <script src="http://apps.bdimg.com/libs/bootstrap/3.2.0/js/bootstrap.min.js"></script>
    <style>
    .text {
   	color:#009900;
   	}
	.des-center {
	    text-align: center;
	    font-size:24px;
	    color:#4876FF;
	    padding-bottom:5px;
	}
	</style>
</head>
<body style="padding-top:55px;">
{{str2html .Navigation}}
<div class="container-fluid">

<div class="row">
<div class="col-md-6 col-md-offset-2">
    <form class="form-horizontal" role="form">
    	<div class="row form-group">
    		<label  class="col-md-2 control-label text">输入账号</label>
    		<div class="col-md-4">
    			<input type="text" class="form-control" id="name" 
    				   placeholder="32个以内中文、字母或数字">
    		</div>
    	</div>
    	<div class="row form-group">
    		<label class="col-md-2 control-label text">输入密码</label>
    		<div class="col-md-4">
    			<input type="password" class="form-control" id="passwd" 
    				   placeholder="6-32位字母或数字">
    		</div>
    	</div>
    	<div class="row form-group">
    	<label class="col-md-2 control-label text">确认密码</label>
    		<div class="col-md-4">
    			<input type="password" class="form-control" id="confirm" 
    				   placeholder="确认密码">
    		</div>
    	</div>
    	<div class="row form-group">
    	<label class="col-md-2 control-label text">你的邮箱</label>
    		<div class="col-md-4">
    			<input type="text" class="form-control" id="mail" 
    				   placeholder="用于找回密码">
    		</div>
    	</div>
    	
    </form>
    <div class="col-md-offset-2">
      <button id="register" type="submit" class="btn btn-default" onclick="doRegister()">注册</button>
      <p id="info" class="alert">网站提示：注册账号！<a href="auth">已有账号？点击登录</a></p>
    </div>
    
    </div>
    <script>
        function check(str){
		    if (str.indexOf(" ") == -1){
		        return true;
		    }
			return false;
        }

		function doRegister() {
	    	var name = document.getElementById("name").value;
			var passwd = document.getElementById("passwd").value;
			var confirm = document.getElementById("confirm").value;
			var mail = document.getElementById("mail").value;
			
			if (name.length < 1) {
				document.getElementById("info").innerHTML=("用户名为空！");
				return 
			} else if(name.length > 32) {
				document.getElementById("info").innerHTML=("用户名长度不能超过32个字符！");
				return 
			} else if (!check(name)) {
				document.getElementById("info").innerHTML=("用户名不能包含空格！");
				return 
			}
			
			if (passwd.length < 6) {
				document.getElementById("info").innerHTML=("密码不能小于6位长度！");
				return 
			} else if(passwd.length > 32) {
				document.getElementById("info").innerHTML=("密码长度不能超过32个字符！");
				return 
			} else if (passwd != confirm) {
				document.getElementById("info").innerHTML=("两次输入密码不一致！");
				return 
			} else if (!check(passwd)) {
				document.getElementById("info").innerHTML=("密码不能包含空格！");
				return 
			}else if(mail.length < 1) {
				document.getElementById("info").innerHTML=("没有输入用于找回密码的邮箱！请输入邮箱");
				return 
			}else if (!check(mail)) {
				document.getElementById("info").innerHTML=("邮箱不能包含空格！");
				return 
			}

			var md5Pwd=$.md5(passwd);

			document.getElementById("info").innerHTML=("正在注册中...");
			var btn = $("#register"); 
			btn.attr("disabled", true);

			$.ajax({
		          async:true,
		          cache:false,
		          timeout:10000,
		          type:"POST",
		          url:"/register",
		          data:{
			       name_:name,
				   passwd_:md5Pwd,
				   mail_:mail,
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
						btn.attr("disabled", false);
						document.getElementById("info").innerHTML=(msg);
		          	} else {
						document.getElementById("info").innerHTML=("注册成功！");
						window.location.href="/"
		          	}
		            
		          }
		        });
		}
</script>
</div>
</div>
</body>
</html>
