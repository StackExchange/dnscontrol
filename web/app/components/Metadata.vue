<template>
    <div style="position: relative">
    <popover v-if="hasMeta()" effect="fade" placement="right" title="Metadata" trigger="hover">
        <span  class="label label-info noselect">Metadata</span>
        <div slot="content">
            <table class='table '>
                <tbody>
                    <tr v-for="(k,v) in meta">
                        <td><b>{{k}}</b></td>
                        <td>{{v}}</td>
                    </tr>
                </tbody>
            </table>
        </div>
    </popover>
    </div>
</template>

<script>
    import { popover } from 'vue-strap'

    export default{
        name: "metadata",
        data: function(){return{
            
        };},
        methods:{
            hasMeta: function(){
                return Object.keys(this.meta).length > 0;
            }
        },
        props:{
            meta:{type:Object, required: true}
        },
        components:{popover}
    }

</script>

<style>
</style>