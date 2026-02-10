from hatchling.builders.hooks.plugin.interface import BuildHookInterface

import pathlib
import shutil

repo_dir = pathlib.Path(__file__).joinpath("../../../../../../src").resolve()


def generate():
    src = pathlib.Path(__file__).joinpath("../src").resolve()
    if src.exists():
        shutil.rmtree(src)
    src.mkdir()

    realsrc = repo_dir.joinpath("proton-rds-sdk-py/dbutilsx")
    shutil.copytree(realsrc, src.joinpath("dbutilsx"))

def reset():
    src = pathlib.Path(__file__).joinpath("../src").resolve()
    if src.exists():
        shutil.rmtree(src)


class BuildFrontend(BuildHookInterface):
    PLUGIN_NAME = "rds-sdk_DBUtilsX"
    def initialize(self, version, build_data) -> None:
        generate()
        super().initialize(version, build_data)

    def finalize(self, version, build_data, artifact_path) -> None:
        super().finalize(version, build_data, artifact_path)
        reset()
