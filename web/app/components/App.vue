<template>
  <div class="container">
    <div class='clearfix'>
        <h3 class='pull-left'> DNSControl </h3>
        <button class='btn btn-success pull-right' @click="save" style='margin-top:15px'>Save</button>
    </div>
    <multiselect name="domains" :options="domains" placeholder="jump to domain" :selected="null" :multiple="false" :searchable="true" :on-change="domainSelect" :show-labels="false" :reset-after="true"></multiselect>
    <br/>
    <multiselect name="vars" :options="vars" placeholder="jump to var" :selected="null" :multiple="false" :searchable="true" :on-change="varSelect" :show-labels="false" :reset-after="true"></multiselect>
    <br/>
    <editor :text.once="content" v-ref:editor :on-change="textChanged"></editor>
    
    <div v-show="error" class="alert alert-danger" role="alert">{{error}}</div>
    <dns-config v-if="config" :data="config" :selected="lastSelected" ></dns-config>
  </div>
</template>

<script>
import Multiselect from 'vue-multiselect'
import Editor from './Editor'
import DnsConfig from './DNSConfig'
import 'whatwg-fetch';
import _ from 'lodash';
window._ = _;

export default {
    data:function(){
        return {
        domains: [],
        vars: [],
        content: window.initialScript,
        config: null,
        error: null,
        lastSelected: ""
        };
    },
    created: function(){
        this.parse();
    },
    methods:{
        parse: function(){
            var domainRE = this.buildDomainRegex();
            var match;
            var domains = [];
            while(match = domainRE.exec(this.content)){
                domains.push(match[2]);
            }
            domains.sort(function (a, b) {return a.toLowerCase().localeCompare(b.toLowerCase());});
            var varRE = this.buildVarRegex();
            var vars = [];
            while(match = varRE.exec(this.content)){
                vars.push(match[1]);
            }
            vars.sort(function (a, b) {return a.toLowerCase().localeCompare(b.toLowerCase());});
            this.domains = domains;
            this.vars = vars;
            this.run();
        },
        textChanged: function(e){
            this.content = e;
            this.parse();
        },
        domainSelect: function(val){
            var searchRegex = this.buildDomainRegex(val);
            this.$refs.editor.jump(searchRegex);
            this.lastSelected = val;
        },
        varSelect: function(val){
            var searchRegex = this.buildVarRegex(val);
            this.$refs.editor.jump(searchRegex);
        },
        buildDomainRegex: function(d){
            // / ^\s*D\s*\(\s*(['"])((?:\\\1|.)*?)\1/gm; //holy crap
            /*
                ^              //start of line
                \s*D\s*\(\s*   // D( with any whitespace
                (['"])         // single or double quote (matching group 1)
                ((?:\\\1|.)*?) //voodoo. Everything that is not the same char as the start quote
                \1             //same quote again

                Turned into string and double escaped slashes below
            */
            if (!d){
                d = "((?:\\\\\\1|.)*?)"
            }
            return new RegExp("^\\s*D\\s*\\(\\s*(['\"])"+d+"\\1","gm");
        },
        buildVarRegex: function(v){
            if (!v){
                v = "([^\\s=]+)"
            }
            return new RegExp("^\s*var "+v,"gm");
        },
        run: function(){
            this.error = null;
            this.config = null;
            try{
                initialize();
                eval(this.content);
                this.config= conf;
                this.stale = false;
            }catch(e){
                this.error = e.toString();
            }
        },
        save: function(){
            fetch('/api/save', {
                method: 'POST',
                body: this.content
            }) //TODO: error handling
        }
    },
    components:{Multiselect,Editor,DnsConfig}
}
</script>