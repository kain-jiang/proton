from hatchling.builders.hooks.plugin.interface import BuildHookInterface
import platform
import pathlib
from sys import version_info
import shutil

repo_dir = pathlib.Path(__file__).joinpath("../../../../../../src").resolve()


def dm_sofile():
    machine = platform.machine()
    major, minor = version_info[:2]
    py_version = f"{major}{minor}"
    so_srcname = f"rdsdriver/dm/dmPython.cpython-{py_version}-{machine}-linux-gnu.so"
    so_src = repo_dir.joinpath(f"proton-rds-sdk-py/{so_srcname}")
    if not so_src.exists():
        dft_so_name = f"rdsdriver/dm/dmPython.cpython-38-{machine}-linux-gnu.so"
        so_src = repo_dir.joinpath(f"proton-rds-sdk-py/{dft_so_name}")
    return so_src


def generate():
    src = pathlib.Path(__file__).joinpath("../src").resolve()
    if src.exists():
        shutil.rmtree(src)
    src.mkdir()

    realsrc = repo_dir.joinpath("proton-rds-sdk-py/rdsdriver")
    shutil.copytree(realsrc, src.joinpath("rdsdriver"))

    if not repo_dir.joinpath(f"proton-rds-sdk-py/rdsdriver/dm").exists():
        return

    for f in src.glob("rdsdriver/dm/*.so"):
        f.unlink()
    shutil.copy2(dm_sofile(), src.joinpath("rdsdriver/dm/dmPython.so"))

def reset():
    src = pathlib.Path(__file__).joinpath("../src").resolve()
    if src.exists():
        shutil.rmtree(src)


class BuildFrontend(BuildHookInterface):
    PLUGIN_NAME = "rds-sdk_RDSDriver"
    def initialize(self, version, build_data) -> None:
        generate()
        super().initialize(version, build_data)

    def finalize(self, version, build_data, artifact_path) -> None:
        super().finalize(version, build_data, artifact_path)
        reset()
