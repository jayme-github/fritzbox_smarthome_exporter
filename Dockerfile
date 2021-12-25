FROM scratch

ARG TARGETOS
ARG TARGETARCH

COPY fritzbox_smarthome_exporter /fritzbox_smarthome_exporter

EXPOSE 9103

ENTRYPOINT ["/fritzbox_smarthome_exporter"]
