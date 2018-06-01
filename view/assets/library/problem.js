$(document)
    .ready(function () {
        $.get('/userinfo', function (data) {
            $('#username').after(data);
        });

        var myCodeMirror = CodeMirror.fromTextArea(document.getElementById("editor"), {
            mode: "cmake",
            theme: "default",
            lineNumbers: true,
            value: "put your code here"
        });

        $('.ui.dropdown').dropdown();

        $('.ui.selection.dropdown').dropdown({
            onChange: function () {
                var lan = $('.ui.dropdown')
                    .dropdown('get value');
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
    });