import React, { FC, useState, useEffect } from "react";
import { DescriptionEditProps } from "./declare";
import { Button, Space } from "@kweaver-ai/ui";
import __ from "../../locale";

function EditComponent<T = any>(props: DescriptionEditProps<T>) {
  const { value, edit, onSave, onCancel, render, Component } = props;

  const isEditControl = () => edit !== undefined;

  const [innerValue, setInnerValue] = useState(value);
  const [innerEdit, setInnerEdit] = useState(isEditControl() ? edit : false);
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    setInnerValue(value);
  }, [value]);

  useEffect(() => {
    if (isEditControl()) {
      setInnerEdit(edit);
    }
  }, [edit]);

  const onValueChange = (value: T) => {
    const v: any = value as any;

    if (typeof v === "object" && v?.target?.value !== undefined) {
      setInnerValue(v.target.value);
    } else {
      setInnerValue(value);
    }
  };

  const changeEdit = (edit: boolean) => {
    if (!isEditControl()) {
      setInnerEdit(edit);
    }
  };
  const onSaveFn = () => {
    setLoading(true);
    if (onSave) {
      const result = onSave(innerValue);

      if (typeof result === "object") {
        result
          .then(() => {
            setLoading(false);
          })
          .catch(() => {
            setLoading(false);
          })
          .finally(() => {
            changeEdit(false);
          });
      }
    } else {
      changeEdit(false);
      setLoading(false);
    }
  };

  const onCancelFn = () => {
    if (onCancel) {
      onCancel();
    }
    setInnerValue(value);
    changeEdit(false);
  };

  const renderContent = () => {
    return render
      ? render(value)
      : typeof value === "string" || typeof value === "number"
      ? value
      : "";
  };

  return (
    <div>
      {innerEdit ? (
        <Space direction="vertical" style={{ width: "100%" }}>
          <div>
            {React.cloneElement(Component, {
              value: innerValue,
              onChange: onValueChange,
            })}
          </div>
          <Space direction="horizontal">
            <Button type="primary" onClick={onSaveFn} loading={loading}>
              {__("确定")}
            </Button>
            <Button onClick={onCancelFn}>{__("取消")}</Button>
          </Space>
        </Space>
      ) : (
        <React.Fragment>{renderContent()}</React.Fragment>
      )}
    </div>
  );
}

export default EditComponent;
