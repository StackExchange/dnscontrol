<template>
    <div class='domain'>
        <div class='row'>
            <div class='col-md-12' style='padding-left:20px;padding-right:25px;'>
                <div class='pull-left'><h2>{{d.name}}</h2></div>
                <div class='meta pull-left'><metadata :meta="d.meta"></metadata></div>
                <tooltip trigger="hover" placement="bottom" content="Compare data with provider state and view corrections">
                    <div class='pull-right'><button class='btn btn-default btn-domain' @click="preview">Preview</button></div>
                </tooltip>
            </div>
        </div>
        <tooltip trigger="hover" placement="bottom" :content="regTooltip()">
            <span :class="['label','noselect',registrarEnabled?'label-success':'label-default']" @click="registrarEnabled = !registrarEnabled">{{d.registrar}}</span>
        </tooltip>
        <tooltip v-for="(i,dsp) in d.dsps" trigger="hover" placement="bottom" :content="dspTooltip(i)">
            <span :class="['label','noselect',dsps[i]?'label-success':'label-default']" @click="dsps.$set(i,!this.dsps[i])">{{dsp}}</span>
        </tooltip>
        <record v-for="record in d.records" :record="record"></record>
    </div>
</template>

<script>
import Metadata from './Metadata'
import Record from './Record'
import { tooltip } from 'vue-strap'

export default{
    name:"domain",
    props: {
        d:Object,
        onPreview: {type:Function, required:true} //accepts a data object and returns a fetch promise
    },
    data:function(){return{
        registrarEnabled: true,
        dsps: []
    }},
    created: function(){
        for(var i = 0; i<this.d.dsps.length; i++){
            this.dsps[i] = true;
        }
    },
    methods:{
        preview: function(){
            var data = {
                Domain: this.d.name,
                Registrar: this.registrarEnabled,
                Dsps: this.dsps
            }
            this.onPreview(data);
        },
        regTooltip: function(){
            return (this.registrarEnabled?"Disable":"Enable")+" checking of registrar entries with "+this.d.registrar;
        },
        dspTooltip: function(i){
            return (this.dsps[i]?"Disable":"Enable")+" checking of domain records with "+this.d.dsps[i];
        }
    },
    components: {Metadata,Record,tooltip}
}
</script>

<style>
.domain{
  padding-bottom: 15px;
}
.noselect {
  -webkit-touch-callout: none; 
  -webkit-user-select: none;   
  -khtml-user-select: none;    
  -moz-user-select: none;      
  -ms-user-select: none;       
  user-select: none;           
  cursor: pointer;
}
.domain:nth-of-type(odd) {
        background: #e0e0e0;
}
.label {
  margin-right: 4px;
}
.btn-domain {
  margin-top: 20px;
}
.meta{
    margin-top: 28px;
    margin-left: 10px;
}
</style>