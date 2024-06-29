import React from 'react';
import { Card } from 'antd';
import { Link } from 'react-router-dom';

const { Meta } = Card;

export function ProjectCard({ project }) {
  return (
    <div className={'ProjectCard_container'}>
      <Card bordered={true} style={{ minWidth: 300 }}>
        <Meta title={<Link title={project.name} to={`/projects/${project.id}`}>{project.name}</Link>} description={project.description} />
      </Card>
    </div>
  );
}
