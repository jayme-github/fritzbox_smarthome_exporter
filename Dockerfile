FROM scratch

ARG TARGETOS
ARG TARGETARCH

COPY build/fritzbox_smarthome_exporter.${TARGETOS}.${TARGETARCH} /fritzbox_smarthome_exporter

EXPOSE 9103

ENTRYPOINT ["/fritzbox_smarthome_exporter"]
