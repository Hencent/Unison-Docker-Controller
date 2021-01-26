# Unison Docker Controller
UDC (Unison Docker Controller) is a local docker controller used to monitor local conditions and control local docker containers.

UDC is an integral part of UMLCP (Unison Machine Learning Cloud Platform).

**衡 · Docker 控制器** 是一个本地 docker 控制器，用于监控本机状况并控制本地 docker 容器。

**衡 · Docker 控制器** 是 **衡 · 机器学习云平台** 的组成部分。

# Special Notes
+ Disk limit is not available now.

[comment]: <> (+ The host **must not** run other docker containers. Otherwise, other containers will be deleted.)

[comment]: <> (+ It is **not recommended** to run with other programs to avoid affecting the resources obtained by the container.)

# system requirements
+ If you want to use disk limitation, the storage driver should be `overlay2`, and the backing file system is `xfs`, and mounted with the `pquota` mount option. .
+ The disk capacity limit is for the rootfs of the container, and the mounted volume is unlimited。

# Licensing
UDC is licensed under the MIT License. See
[LICENSE](https://github.com/Hencent/Unison-Docker-Controller/blob/main/LICENSE) for the full
license text.