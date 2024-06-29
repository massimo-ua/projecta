import { useParams } from 'react-router-dom';

export function Categories() {
  const { projectId} = useParams();
  return (
    <div>
      Categories {projectId}
    </div>
  );
}
