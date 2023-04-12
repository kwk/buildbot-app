#!/bin/bash 

# This script shows information about the worker in a machine-consumable JSON
# output.

# You can pass --json to this script to output the information in JSON format.

(
    # General
    echo "system_information|$(uname -a)";
    echo "cpu_model|$(cat /proc/cpuinfo | grep "model name" | head -n1 | cut -d " " -f 3-)";
    echo "num_cpu_cores|$(nproc)";
    echo "operating_system|$(lsb_release -d | cut -f 2-)";
    echo "bash_version|$(bash --version | head -n1)";
    echo "architecture|$(arch)";
    
    # Compilers
    [ -x "$(command -v gcc)" ] && echo "gcc_version|$(gcc --version | head -n1)";
    [ -x "$(command -v clang)" ] && echo "clang_version|$(clang --version | head -n1)";
    [ -x "$(command -v ccache)" ] && echo "ccache_version|$(ccache --version | head -n1)";
    [ -x "$(command -v distcc)" ] && echo "distcc_version|$(distcc --version | head -n1)";
    [ -x "$(command -v go)" ] && echo "go_version|$(go version)";
    
    # Debuggers
    [ -x "$(command -v gdb)" ] && echo "gdb_version|$(gdb --version | head -n1)";
    [ -x "$(command -v lldb)" ] && echo "lldb_version|$(lldb --version | head -n1)";
    
    # CI tools
    [ -x "$(command -v buildbot-worker)" ] && echo "buildbot_worker_version|$(buildbot-worker --version | head -n1 | tr -c -d '[0-9.]')";
    [ -x "$(command -v buildslave)" ] && echo "buildslave_version|$(buildslave --version | head -n1 | tr -c -d '[0-9.]')";
    [ -x "$(command -v buildkite-agent)" ] && echo "buildkite_agent_version|$(buildkite-agent --version | head -n1)";
    
    # Linkers
    echo "ld_version|$(ld --version | head -n1)";
    [ -x "$(command -v ld.lld)" ] && echo "ldd_version|$(ld.lld --version | head -n1)";
    [ -x "$(command -v ld.gold)" ] && echo "gold_version|$(ld.gold --version | head -n1)";
    
    # Python
    [ -x "$(command -v python3)" ] && echo "python_version|$(python3 --version | tr -d '[:alpha:][:blank:]')";
    [ -x "$(command -v pip)" ] && echo "pip_version|$(pip --version)";
    [ -x "$(command -v swig)" ] && echo "swig_version|$(swig -version | head -n2 | tr -c -d '[0-9.]')";
    
    # Configure/CMake
    [ -x "$(command -v autoconf)" ] && echo "autoconf_version|$(autoconf --version | head -n1 | tr -c -d '[0-9.]')";
    echo "cmake_version|$(cmake --version | head -n1 | tr -d '[:alpha:][:blank:]')";
    [ -x "$(command -v make)" ] && echo "make_version|$(make --version | head -n1)";
    [ -x "$(command -v ninja)" ] && echo "ninja_version|$(ninja --version | head -n1)";
    
    # Other
    echo "git_version|$(git --version | head -n1 | tr -d '[:alpha:][:blank:]')";

    # GPU stuff
    [ -x "$(command -v vulkaninfo)" ] && echo "vulkan_instance_version|$(vulkaninfo 2>/dev/null | grep "Vulkan Instance" | cut -d " " -f 4-)"; 
    [ -x "$(command -v vulkaninfo)" ] && echo "nvidia_vulkan_icd_version|$(vulkaninfo 2>/dev/null | grep "apiVersion" | cut -d= -f2 | awk '{printf $2}' | tr -d '()')";
) | column \
    -s '|' \
    -t \
    --table-name "worker_information" \
    --table-columns "key,value" \
    -o "  " \
    "$@"
