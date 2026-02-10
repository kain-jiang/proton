import * as React from "react";
import { Row, Col } from "@aishutech/ui";
import { DeleteOutlined, QuestionCircleOutlined } from "@aishutech/ui/icons";
import { Props } from "./index.d";

export const Title: React.FC<Props> = React.memo(
  ({ title, tip, deleteCallback }) => {
    return (
      <>
        <Row>
          <Col
            span={deleteCallback ? 23 : 24}
            style={{
              color: "#000000",
              height: "30px",
              lineHeight: "30px",
              fontSize: "14px",
              fontWeight: "bold",
            }}
          >
            <span className="split"></span>
            <span>{title}</span>
            {tip ? (
              <QuestionCircleOutlined
                style={{
                  marginLeft: "6px",
                }}
                title={tip}
              />
            ) : null}
          </Col>
          {deleteCallback ? (
            <Col className="delete" span={1}>
              <DeleteOutlined onClick={deleteCallback} />
            </Col>
          ) : null}
        </Row>
        <div
          style={{
            borderTop: "2px solid #EEEEEE",
            margin: "10px 0",
          }}
        ></div>
      </>
    );
  },
);
