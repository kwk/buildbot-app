#!/bin/bash

set -eu

# Ensure Bash pipelines (e.g. cmd | othercmd) return a non-zero status if any of
# the commands fail, rather than returning the exit status of the last command
# in the pipeline.
set -o pipefail

# We installed buildbot in a virtualenv that we need to activate here
source /home/bb-master/sandbox/bin/activate

BUILDBOT_MASTER_BASEDIR=/home/bb-master/basedir
CONFIG_FILE=/home/bb-master/cfg/master.cfg

buildbot create-master --force --config=${CONFIG_FILE} "${BUILDBOT_MASTER_BASEDIR}"

buildbot start "${BUILDBOT_MASTER_BASEDIR}" || (tail ${BUILDBOT_MASTER_BASEDIR}/twistd.log && exit 3)

echo "Monitoring file changes to ${CONFIG_FILE} in the background"
while true; do
    inotifywait -q --event modify --format '%w' ${CONFIG_FILE}
    echo "Detected change in ${CONFIG_FILE}. Sending SIGHUP to buildbot master to re-read the config."
    # http://docs.buildbot.net/latest/manual/cmdline.html#sighup
    buildbot sighup ${BUILDBOT_MASTER_BASEDIR}
done &

echo "Following master log..."
tail -f ${BUILDBOT_MASTER_BASEDIR}/twistd.log