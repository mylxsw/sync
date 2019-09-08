<template>
    <b-row class="mb-5">
        <b-col>
            <b-progress :value="progress" show-progress animated class="mb-2"></b-progress>
            <b-card>
                <b-card-body body-bg-variant="dark" body-text-variant="white">
                    <b-card-text v-for="(line, index2) in messages" :key="index2">
                        <b-badge class="mr-2">{{ line.index }}</b-badge>
                        <b class="text-success">{{ line.timestamp }}</b>
                        <b-badge class="ml-2" :variant="line.level === 'ERROR' ? 'danger' : 'info'">{{ line.level }}
                        </b-badge>
                        <br/>
                        {{ line.message }}
                    </b-card-text>
                </b-card-body>
            </b-card>
        </b-col>
    </b-row>
</template>

<script>
    export default {
        name: 'Job',
        data() {
            return {
                messages: [],
                errors: [],
                progress: 0,
            };
        },
        mounted() {
            let self = this;

            let conn = new WebSocket(this.$store.getters.websocketUrl + "/api/running-jobs/" + this.$route.params.id + "/");
            conn.onclose = function (evt) {
                console.log(evt);
            };
            conn.onerror = function (evt) {
                console.log(evt);
            };
            conn.onmessage = function (evt) {
                let data = JSON.parse(evt.data);
                let payload = JSON.parse(data.payload);
                switch (data.type) {
                    case "console":
                        if (self.messages.length > 20) {
                            self.messages.shift();
                        }
                        self.messages.push(payload.console);
                        if (payload.level === 'ERROR') {
                            self.errors.push(payload.console);
                        }
                        break;
                    case "progress":
                        self.progress = payload.percentage * 100;
                        break;
                    default:
                }
            };
        }
    }
</script>
