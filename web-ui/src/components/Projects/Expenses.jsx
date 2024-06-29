import { useParams } from 'react-router-dom';

export function Expenses() {
  const { projectId} = useParams();
  return (
    <div>
      Expenses {projectId}
    </div>
  );
}
