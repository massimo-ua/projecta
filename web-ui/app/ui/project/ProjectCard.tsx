import React from 'react';
import { Project } from '@/app/ui/project/types';

export default function ProjectCard(props: Project) {
  return (
    <div className="flex flex-col bg-white rounded-md shadow-md p-6">
      <h2 className="text-xl font-semibold">{props.name}</h2>
      <p className="text-gray-500 text-sm">{props.description}</p>
      <div className="flex flex-row justify-between items-center mt-4">
        <button className="flex items-center gap-2 bg-blue-600 text-white rounded-md p-2">
          <p className="text-sm font-semibold">View</p>
        </button>
      </div>
    </div>
  );
}
