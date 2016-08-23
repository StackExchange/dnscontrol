<template>
    <div>
        <div class='clearfix'>
            <div class='pull-left'>{{correction.Msg}} 
                <span v-show="!hasRun" class="label label-primary noselect" @click="run">Run!</span>
                <span v-show="ok" class='glyphicon glyphicon-ok' style='color:rgb(93, 197, 150);'></span>
            </div>
            <spinner :loading="running" class='pull-left' height="15px" ></spinner>
        </div>
        <alert v-if="error" type="danger" >
            {{error}}
        </alert>
    </div>
</template>
<script>
    import spinner from 'vue-spinner/src/ScaleLoader.vue'
    import { alert } from 'vue-strap'

    export default{
        data: function(){return{
            hasRun: false,
            running: false,
            error: "",
            ok: false
        };},
        methods: {
            run: function(){
                if (this.hasRun){return;}
                this.hasRun = true;
                this.running = true;
                this.error = "";
                this.ok = false;
                var self = this;
                fetch("/api/run?id="+this.correction.ID,{method: "POST"})
                .then(function(response) {
                    if (response.status != 200){
                        response.text().then(function(txt){
                            self.error = txt;
                            self.running = false;
                            self.hasRun = false;
                        })
                        return
                    }
                    self.ok= true
                    self.running=false
                });
            }
        },
        props:{
            correction: Object
        },
        components:{spinner,alert}
    }
</script>

<style>
</style>