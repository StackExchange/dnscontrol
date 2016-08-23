<template>
    <modal :show.sync="show" :title="query.Domain" width="75%">
        <div slot='modal-body'><div style='margin:15px;'>
            <spinner :loading="loading"></spinner>
            <alert v-if="error" type="danger" >
                {{error}}
            </alert>
            <div v-for="(provider,list) in corrections">
                <h3>{{provider}} <span v-show="list.length==0" class='glyphicon glyphicon-ok' style='color:rgb(93, 197, 150);'></span></h3>
                <div v-for="correction in list">
                    <correction :correction="correction"></correction>
                </div>
            </div>
        </div></div>
        <div slot="modal-footer" class="modal-footer">
            <button type="button" class="btn btn-default" @click="show = false">Close</button>
        </div>
    </modal>
</template>

<script>
    import { modal,alert } from 'vue-strap'
    import spinner from 'vue-spinner/src/ScaleLoader.vue'
    import Correction from './Correction'
    export default{
        props:{
            show:{
                required: true,
                type: Boolean,
                twoWay: true
            },
            query: {type: Object, default:{}},
            config: Object
        },
        data: function(){return{
            loading: false,
            error: "",
            corrections: []
        };},
        methods:{
            preview: function(){
                var self = this;
                fetch('/api/preview', {
                    method: 'POST',
                    headers: {
                        'Accept': 'application/json',
                        'Content-Type': 'application/json'
                    },
                    body: JSON.stringify({Config: this.config, Query: this.query})
                }).then(function(response) {
                    if (response.status != 200){
                        response.text().then(function(txt){
                            self.error = txt;
                            self.loading = false;
                        })
                        return
                    }
                    response.json().then(function(j){
                        self.corrections= j
                        self.loading=false
                    })
                });
            }
        },
        created: function(){
            this.$watch('show', function(newVal){
                if (newVal && !this.loading){
                    this.loading = true;
                    this.error = "";
                    this.corrections = [];
                    this.preview();
                    
                }
            })
        },
        components:{modal,alert,spinner,Correction}
    }
</script>

<style>
</style>