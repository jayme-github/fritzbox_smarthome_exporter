FROM scratch

ARG TARGETOS
ARG TARGETARCH

COPY passwd /etc/passwd
COPY fritzbox_smarthome_exporter /fritzbox_smarthome_exporter

EXPOSE 9103

USER 65534
ENTRYPOINT ["/fritzbox_smarthome_exporter"]
