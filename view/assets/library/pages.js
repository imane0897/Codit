$(document).ready(function () {
    $('.ui.dropdown').dropdown();

    $.get('/userinfo', function (data) {
        $('#username').after(data);
    });

    $.get('/isadmin', function (data) {
        if (data == "ture") {
            var txt = "<div class=\"item \"> <a href = \"dashboard\" style = \"color:black\" > <i class = \"edit icon\" > </i>Dashboard</a> </div>";
            $('#setting').after(txt);
        }
    });
});