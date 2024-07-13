import { Col, Row } from 'antd';

export default function Footer() {
  const startYear = 2024;
  const currentYear = new Date().getFullYear();
  const devPeriod = startYear === currentYear ? currentYear : `${startYear}-${currentYear}`;
  return (<Row>
    <Col span={24} style={{ textAlign: 'center' }}>
      {`Web UI ©${devPeriod} Created by Massimo UA`}
    </Col>
  </Row>);
}
