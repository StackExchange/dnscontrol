<template>
    <div class='clearfix'><b class='pull-left'>{{record.type}} {{record.name}} {{record.target | truncate 50}}</b> 
        <div style='margin-left:8px' class='pull-left'><metadata :meta="record.meta"></metadata></div></div>
</template>
<script>
    import Metadata from './Metadata'
    export default{
        data: function(){return{
            
        };},
        props: {
            record:{type: Object, required:true}
        },
        components:{Metadata}
    }
</script>

<style>
</style>