$(document)
    .ready(function () {
        $.get('/pidcount', function (data) {
            len = parseInt(data);
            for (var i = 0; i < len; i++) {
                $('#pidlist').append('<div class="item">' + (1000 + i) + '</div>');
            }
            newIndex = len + 1000;
            $('#pidlist').append('<div class="item">' + newIndex + '</div>');
        });

        $('.ui.dropdown').dropdown();

        $('.ui.fluid.search.selection.dropdown').dropdown({
            onChange: function (val) {
                if ($('.ui.dropdown').has('#pid').dropdown('get value') == newIndex.toString()) {
                    $('form').attr('action', '/newproblem');
                } else {
                    $.getJSON('/editproblem?pid=' + escape(val), function (data) {
                        $('#title').val(data.Title)
                        $('.ui.dropdown').has('#level').dropdown('set selected', data.Level.toString());
                        $('#description').val(data.Description);
                        $('#input').val(data.Input);
                        $('#output').val(data.Output);
                        $('#sampleinput').val(data.SampleInput);
                        $('#sampleoutput').val(data.SampleOutput);
                    });
                }
            }
        });

        $('button.primary').click(function () {
            $('.ui.dropdown').has('#pid').addClass("disabled");
            $('form').attr('action', '/newproblem');
            // clear form
            $('#title').val('');
            $('.ui.dropdown').has('#level').dropdown('clear');
            $('#description').val('');
            $('#input').val('');
            $('#output').val('');
            $('#sampleinput').val('');
            $('#sampleoutput').val('');
            $('.ui.dropdown').has('#pid').dropdown('set selected', newIndex.toString());
        });

        $('button.secondary').click(function () {
            $('.ui.dropdown').has('#pid').removeClass("disabled");
            // clear form
            $('.ui.dropdown').has('#pid').dropdown('set selected', newIndex.toString());
            $('#title').val('');
            $('.ui.dropdown').has('#level').dropdown('clear');
            $('#description').val('');
            $('#input').val('');
            $('#output').val('');
            $('#sampleinput').val('');
            $('#sampleoutput').val('');
        });

        $('.ui.form').form({
            fields: {
                title: {
                    identifier: 'title',
                    rules: [{
                        type: 'empty',
                        prompt: 'Please enter question title'
                    }]
                },
                level: {
                    identifier: 'level',
                    rules: [{
                        type: 'empty',
                        prompt: 'Please select a level for the question'
                    }]
                },
                description: {
                    identifier: 'description',
                    rules: [{
                        type: 'empty',
                        prompt: 'Please enter question description'
                    }]
                },
                input: {
                    identifier: 'input',
                    rules: [{
                        type: 'empty',
                        prompt: 'Please enter question input description'
                    }]
                },
                output: {
                    identifier: 'output',
                    rules: [{
                        type: 'empty',
                        prompt: 'Please enter question output description'
                    }]
                },
                sampleinput: {
                    identifier: 'sampleinput',
                    rules: [{
                        type: 'empty',
                        prompt: 'Please enter question sample input'
                    }]
                },
                sampleoutput: {
                    identifier: 'sampleoutput',
                    rules: [{
                        type: 'empty',
                        prompt: 'Please enter question sample output'
                    }]
                }
            }
        });
    });