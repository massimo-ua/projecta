import React from 'react';
import { Card, Typography } from 'antd';
import { Link } from 'react-router-dom';

const { Meta } = Card;
const { Paragraph } = Typography;

export function ProjectCard({ project }) {
  return (
    <div className="ProjectCard_container">
      <Card bordered style={{ height: '100%', display: 'flex', flexDirection: 'column' }}>
        <Meta
          title={
            <Link title={project.name} to={`/projects/${project.id}`}>
              <Paragraph ellipsis={{ rows: 1 }} style={{ marginBottom: 0 }}>
                {project.name}
              </Paragraph>
            </Link>
          }
          description={
            <Paragraph ellipsis={{ rows: 2 }} type="secondary">
              {project.description}
            </Paragraph>
          }
        />
      </Card>
    </div>
  );
}
