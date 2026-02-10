from hatchling.builders.hooks.plugin.interface import BuildHookInterface
import platform
import pathlib
import os
import shutil

repo_dir = pathlib.Path(__file__).joinpath("../../../../../repos").resolve()

def fetch_bin() -> pathlib.Path:
    bin_name = "thrift"
    if platform.system() == "Windows":
        bin_name = "thrift.exe"
    elif platform.machine() == 'aarch64':
        bin_name = 'thrift-arm'
    else:
        bin_name = "thrift"
    
    return repo_dir.joinpath("API/ThriftAPI").joinpath(bin_name).resolve()
    

def generate():
    src = repo_dir.joinpath("API/ThriftAPI").resolve()
    dst = pathlib.Path(__file__).joinpath("../src").resolve()
    if dst.exists():
        shutil.rmtree(dst)
    dst.mkdir(parents=True, exist_ok=True)
    bin = fetch_bin()

    for f in src.iterdir():
        if f.is_file() and f.name.endswith(".thrift"):
            cmd = f"{str(bin)} -r --gen py --out {str(dst)} {str(f)}"
            os.system(cmd)

def reset():
    src = pathlib.Path(__file__).joinpath("../src").resolve()
    if src.exists():
        shutil.rmtree(src)


class BuildFrontend(BuildHookInterface):
    PLUGIN_NAME = "ThriftAPI"
    def initialize(self, version, build_data) -> None:
        generate()
        super().initialize(version, build_data)

    def finalize(self, version, build_data, artifact_path) -> None:
        super().finalize(version, build_data, artifact_path)
        reset()