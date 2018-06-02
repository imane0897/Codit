$(document)
    .ready(function () {

        var myCodeMirror = CodeMirror.fromTextArea(document.getElementById("editor"), {
            mode: "cmake",
            theme: "default",
            lineNumbers: true,
            value: "put your code here"
        });

        $('.ui.selection.dropdown').dropdown('setting', 'onChange', function () {
            var lan = $('.ui.dropdown').dropdown('get value');
            if (lan == "c") {
                myCodeMirror.getDoc().setValue(
                    "#include <stdio.h> \nint main(void) {\n\tprintf(\"Hello World!\\n\");\n\treturn 0;\n}");
            } else if (lan == "cpp") {
                myCodeMirror.getDoc().setValue(
                    "#include <iostream>\nusing namespace std;\nint main() {\n\tcout << \"Hello World!\" << endl;\n\treturn 0;\n}"
                );
            } else if (lan == "java") {
                myCodeMirror.getDoc().setValue(
                    "class myjavaprog \n{\n\tpublic static void main(String args[])\n\t{\n\tSystem.out.println(\"Hello World!\");\n\t}\n}"
                );
            }
        });

        $('.ui.form').form({
            fields: {
                compiler: {
                    identifier: 'compiler',
                    rules: [{
                        type: 'empty',
                        prompt: 'Please select a compiler'
                    }]
                },
                code: {
                    identifier: 'code',
                    rules: [{
                        type: 'empty',
                        prompt: 'Code area cannot be empty'
                    }]
                }
            }
        });

        $('#code-submit').submit(function (event) {
            $('#result').remove();
            $(this).after("<br>" +
                    "<div id=\"result\">"+
                    "<h3>Submission Result: " +
                    "<span id=\"res\" class=\"pending\">Pending</span>" +
                    "<div class=\"ui tiny active inline loader\"></div>"+
                    "</h3>"+
                    "</div>");
            var url = "/submit";
            $.ajax({
                type: "POST",
                url: url,
                data: $(this).serialize(),
                success: function (data) {
                    $('#res').removeClass('pending');
                    var obj = jQuery.parseJSON(data);
                    var res = obj.Result;
                    alert(obj.Result);
                    if (res == "1") {
                        $('#res').addClass('accept');
                        $('#res').text('Accept');
                    } else {
                        $('#res').addClass('error');
                        if (res == "2") {
                            $('#res').text('Wrong Answer');
                        } else if (res == "3") {
                            $('#res').text('Compile Error');
                        } else if (res == "4") {
                            $('#res').text('Runtime Error');
                        } else if (res == "5") {
                            $('#res').text('Time Limit Exceeded');
                        } else {
                            $('#res').text('System Error');
                        }
                    }
                    $('.ui.tiny.active.inline.loader').remove();
                }
            });
            event.preventDefault();
        });
    });

