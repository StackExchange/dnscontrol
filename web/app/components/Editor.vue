<template>
    <pre class="editor" v-el:editor>{{text}}</pre>
</template>

<script>

import ace from 'brace'
import 'brace/mode/javascript';
import 'brace/ext/searchbox';
import 'brace/ext/language_tools';

export default {
    props: {
        text: {type: String, required: true},
        onChange: {type: Function}
    },
    data: function(){ return{
            editor: null,
            debounced: null,
        }
    },
    methods:{
        change: function(e){
            if (!this.debounced){
                var self = this;
                this.debounced = _.debounce(
                    function(){
                        if (self.onChange){
                            self.onChange(self.editor.getValue());
                        }
                    },500
                )
            }
            this.debounced();
        },
        jump: function(r){
            this.editor.find(r,{
				backwards: false,
				wrap: true,
				caseSensitive: false,
				wholeWord: false,
				regExp: true,
		    });
        }
    },
    ready: function(){
        var el = this.$els.editor;
        this.editor = ace.edit(el);
        this.editor.getSession().setMode('ace/mode/javascript');
        this.editor.getSession().on('change', this.change);
        this.editor.setOptions({
            enableBasicAutocompletion: true
        });
    }
}

</script>

<style>
    .editor{
        width: 100%;
        height: 500px;
    }
</style>