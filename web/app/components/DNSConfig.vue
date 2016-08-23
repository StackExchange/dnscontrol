<template>
    <div>
        <run-results :show.sync="showModal" :config="data" :query="query"></run-results>
        <domain v-for="domain in sortedDomains" :d="domain" :on-preview="preview"></domain>
    </div>
</template>
<script>
    import Domain from './Domain'
    import RunResults from './RunResults'

    export default{
        data: function(){return{
            showModal: false,
            query: {},
        };},
        props: {
            data: Object,
            selected: String,
        },
        computed: {
            // hack to bring selected domain to the top. 
            // computed property with domains in the correct sort order.
            sortedDomains: function(){
                var selected = this.selected;
                return _.sortBy(this.data.domains,function(d){
                    return d.name != selected;
                },function(d,i){return i;})
            }
        },
        methods:{
            preview: function(data){
                this.query = data;
                this.showModal = true;
            }
        },
        components: {Domain,RunResults}
    }
</script>
<style>
    
</style>