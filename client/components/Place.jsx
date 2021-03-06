import React from 'react';
import PropTypes from 'prop-types';
import { Link } from 'react-router';
import PlaceDate from './PlaceDate';

const Place = ({ editUrl, name, lastVisited, lastSkipped, visitCount, skipCount }) => (
  <tr>
    <td>
      <Link to={editUrl}>
        {name}
      </Link>
    </td>
    <td>
      <PlaceDate
        date={lastVisited}
        defaultString="Never"
      />
    </td>
    <td>
      {visitCount}
    </td>
    <td>
      <PlaceDate
        date={lastSkipped}
        defaultString="Never"
      />
    </td>
    <td>
      {skipCount}
    </td>
  </tr>
);

Place.propTypes = {
  editUrl: PropTypes.string.isRequired,
  name: PropTypes.string.isRequired,
  lastVisited: PropTypes.instanceOf(Date),
  lastSkipped: PropTypes.instanceOf(Date),
  skipCount: PropTypes.number.isRequired,
  visitCount: PropTypes.number.isRequired,
};

Place.defaultProps = {
  lastVisited: undefined,
  lastSkipped: undefined,
};

export default Place;
