from dataclasses import dataclass
import enum
from datetime import datetime
import uuid

from src.lib.db.db_connector import get_db_operate_obj


class UpgradationStatusEnum(enum.Enum):
    Start = "upgrade_start"
    Running = "upgrade_running"
    Failed = "upgrade_failed"
    Success = "upgrade_success"


class UpgradationTypeEnum(enum.Enum):
    MicroService = "micro-service"
    ModularService = "modular-service"


@dataclass
class UpgradationStatusModel(object):
    upgrade_id: str
    name: str
    type: UpgradationTypeEnum
    status: UpgradationStatusEnum
    start: datetime


@dataclass
class UpgradationStatusRecordModel(object):
    upgrade_id: str
    time: datetime
    message: str


class UpgradationStatus(object):
    record_table = "upgradation_status_records"
    status_table = "upgradation_status"
    time_fmt = "%Y-%m-%d %H:%M:%S"

    @classmethod
    def create_upgrade_status(cls, name: str, type: UpgradationTypeEnum) -> str:
        upgrade_id = str(uuid.uuid4())
        now = datetime.now()
        status = UpgradationStatusEnum.Start
        db_operator = get_db_operate_obj()
        db_operator.insert(cls.status_table, {
            "upgrade_id": upgrade_id,
            "name": name,
            "type": type.value,
            "status": status.value,
            "start": now.strftime(cls.time_fmt)
        })
        return upgrade_id

    @classmethod
    def update_status(cls, upgrade_id: str, status: UpgradationStatusEnum):
        sql = f"UPDATE {cls.status_table} SET status = %s WHERE upgrade_id = %s"
        get_db_operate_obj().update(sql, status.value, upgrade_id)

    @classmethod
    def find_all(cls, name: str = "", service_type: UpgradationTypeEnum = None) -> list[UpgradationStatusModel]:
        db_operator = get_db_operate_obj()
        if name and not service_type:
            result: list[dict] = db_operator.fetch_all_result(f"SELECT * FROM {cls.status_table} WHERE name = %s", name)
        elif not name and service_type:
            result: list[dict] = db_operator.fetch_all_result(f"SELECT * FROM {cls.status_table} WHERE type = %s", service_type.value)
        elif not name and not service_type:
            result: list[dict] = db_operator.fetch_all_result(f"SELECT * FROM {cls.status_table} WHERE 1=1")
        else:
            result: list[dict] = db_operator.fetch_all_result(
                f"SELECT * FROM {cls.status_table} WHERE name = %s AND type = %s", name, service_type.value
            )

        return [
            UpgradationStatusModel(
                upgrade_id=r["upgrade_id"],
                name=r["name"],
                type=UpgradationTypeEnum(r["type"]),
                status=UpgradationStatusEnum(r["status"]),
                start=datetime.strptime(r["start"], cls.time_fmt) if isinstance(r["start"], str) else r["start"]
            )
            for r in result
        ]

    @classmethod
    def find_first(
        cls,
        upgrade_id: str = "",
        name: str = "",
        service_type: UpgradationTypeEnum = None,
    ) -> UpgradationStatusModel:
        db_operator = get_db_operate_obj()
        if upgrade_id:
            result: dict = db_operator.fetch_one_result(f"SELECT * FROM {cls.status_table} WHERE upgrade_id = %s LIMIT 1", upgrade_id)
        else:
            if name and not service_type:
                result: dict = db_operator.fetch_one_result(
                    f"SELECT * FROM {cls.status_table} WHERE name = %s ORDER BY start DESC LIMIT 1", name
                )
            elif not name and service_type:
                result: dict = db_operator.fetch_one_result(
                    f"SELECT * FROM {cls.status_table} WHERE type = %s ORDER BY start DESC LIMIT 1", service_type.value
                )
            elif not name and not service_type:
                result: dict = db_operator.fetch_one_result(
                    f"SELECT * FROM {cls.status_table} WHERE 1=1 ORDER BY start DESC LIMIT 1"
                )
            else:
                result: dict = db_operator.fetch_one_result(
                    f"SELECT * FROM {cls.status_table} WHERE name = %s AND type = %s ORDER BY start DESC LIMIT 1",
                    name, service_type.value
                )
        return UpgradationStatusModel(
            upgrade_id=result["upgrade_id"],
            name=result["name"],
            type=UpgradationTypeEnum(result["type"]),
            status=UpgradationStatusEnum(result["status"]),
            start=datetime.strptime(result["start"], cls.time_fmt) if isinstance(result["start"], str) else result[
                "start"]
        ) if result else None

    @classmethod
    def add_record(cls, upgrade_id: str, message: str):
        now = datetime.now()
        if len(message) > 2000:
            # 长度过长，截断信息
            message = f"{message[:1000]}...{message[-1000:]}"
        get_db_operate_obj().insert(cls.record_table, {
            "upgrade_id": upgrade_id,
            "time": now.strftime(cls.time_fmt),
            "message": message
        })

    @classmethod
    def find_records_by_upgrade_id(cls, upgrade_id: str) -> list[UpgradationStatusRecordModel]:
        sql = f"SELECT * FROM {cls.record_table} WHERE upgrade_id = %s"
        result: list[dict] = get_db_operate_obj().fetch_all_result(sql, upgrade_id)
        return [
            UpgradationStatusRecordModel(
                upgrade_id=r["upgrade_id"],
                time=datetime.strptime(r["time"], cls.time_fmt) if isinstance(r["time"], str) else r["time"],
                message=r["message"]
            )
            for r in result
        ] if result else []
