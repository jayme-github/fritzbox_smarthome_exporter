FROM scratch

COPY build/fritzbox_smarthome_exporter.linux.amd64 /

EXPOSE 9103

ENTRYPOINT ["/fritzbox_smarthome_exporter.linux.amd64"]
