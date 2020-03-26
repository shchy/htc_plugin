docker run \
  -v /var/run/docker.sock:/var/run/docker.sock \
  -v $PWD/mackerel-agent.conf:/etc/mackerel-agent/mackerel-agent.conf \
  -v $PWD/conf.d:/etc/mackerel-agent/conf.d:ro \
  -v $PWD/htc_plugin:/etc/mackerel-agent/htc_plugin \
  --name mackerel-agent2 \
  -d \
  mackerel/mackerel-agent


# docker run -v /var/run/docker.sock:/var/run/docker.sock -v %CD%/mackerel-agent.conf:/etc/mackerel-agent/mackerel-agent.conf -v %CD%/conf.d:/etc/mackerel-agent/conf.d:ro -v %CD%/htc_plugin:/etc/mackerel-agent/htc_plugin --name mackerel-agent2 -d mackerel/mackerel-agent

#   -v $PWD/var/lib/mackerel-agent/:/var/lib/mackerel-agent/ \