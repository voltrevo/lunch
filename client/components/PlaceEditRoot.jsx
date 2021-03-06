import React from 'react';
import PropTypes from 'prop-types';
import { Link } from 'react-router';
import EditPlace from '../containers/EditPlace';
import ViewPlace from '../containers/ViewPlace';

function PlaceEditRoot({ params }) {
  return (
    <div className="row">
      <nav className="col-sm-3 col-md-2 hidden-xs-down bg-faded sidebar">
        <ul className="nav nav-pills flex-column">
          <li className="nav-item">
            <Link className="nav-link" to="/manage">
              Places
            </Link>
          </li>
        </ul>
      </nav>

      <main className="col-sm-9 offset-sm-3 col-md-10 offset-md-2 pt-3">
        <h1>Place</h1>
        <div className="table-responsive">
          <EditPlace placeId={params.id} />
          <ViewPlace placeId={params.id} />
        </div>
      </main>
    </div>
  );
}

PlaceEditRoot.propTypes = {
  params: PropTypes.shape({
    id: PropTypes.string.isRequired,
  }).isRequired,
};

export default PlaceEditRoot;
