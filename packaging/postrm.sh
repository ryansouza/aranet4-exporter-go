#!/bin/sh
set -e
if [ -d /run/systemd/system ]; then
  systemctl --system daemon-reload >/dev/null || true
fi

if [ -x "/usr/bin/deb-systemd-helper" ]; then
  if [ "$1" = "remove" ]; then
    deb-systemd-helper mask 'aranet4-exporter.service' >/dev/null || true
  fi

  if [ "$1" = "purge" ]; then
    deb-systemd-helper purge 'aranet4-exporter.service' >/dev/null || true
    deb-systemd-helper unmask 'aranet4-exporter.service' >/dev/null || true
    rm -rf /var/lib/aranet4-exporter /var/cache/aranet4-exporter
  fi
fi
